package meta_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/2785/warframe-assistant/internal/meta"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		dockerHost = "localhost"
	}

	postgres, err := pool.Run("postgres", "13.2-alpine", []string{"POSTGRES_PASSWORD=password"})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		conn := fmt.Sprintf("host=%s port=%s user=postgres password=password dbname=postgres sslmode=disable", dockerHost, postgres.GetPort("5432/tcp"))
		db, err = sqlx.Open("postgres", conn)
		if err != nil {
			fmt.Printf("conn err: %s\n", err)
			return err
		}
		err = db.Ping()
		fmt.Printf("ping err: %s\n", err)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to postgres docker container: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(postgres); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestIGNCrud(t *testing.T) {
	db.MustExec(`
	CREATE TABLE ign (
	    id text NOT NULL PRIMARY KEY,
    	ign text NOT NULL
	);
	`)

	s := &meta.PostgresService{DB: db, Logger: zap.NewNop(), IGNTable: "ign"}

	assert := assert.New(t)

	// test create user
	err := s.CreateIGN("test-user-1", "test-ign-1")
	assert.NoError(err)
	err = s.CreateIGN("test-user-2", "test-ign-2")
	assert.NoError(err)

	// see if we can get these users
	ign, err := s.GetIGN("test-user-1")
	assert.NoError(err)
	assert.Equal("test-ign-1", ign)

	// see if we can list all users
	allIGNs, err := s.ListAllIGN()
	assert.NoError(err)
	assert.EqualValues(
		map[string]string{"test-user-1": "test-ign-1", "test-user-2": "test-ign-2"},
		allIGNs,
	)

	// lets update user1's ign
	err = s.UpdateIGN("test-user-1", "new-ign-1")
	assert.NoError(err)

	// make sure it's changed
	ign, err = s.GetIGN("test-user-1")
	assert.NoError(err)
	assert.Equal("new-ign-1", ign)

	// try deleting one
	err = s.DeleteRelation("test-user-2")
	assert.NoError(err)

	allIGNs, err = s.ListAllIGN()
	assert.NoError(err)
	assert.Equal(1, len(allIGNs))
}

func TestEventCrud(t *testing.T) {
	db.MustExec(`
	CREATE TABLE events (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		guild_id text NOT NULL,
		name text NOT NULL,
		start_date timestamptz DEFAULT current_timestamp,
		end_date timestamptz NOT NULL,
		active boolean,
		event_type text
	);
	`)

	assert := assert.New(t)
	require := require.New(t)

	s := &meta.PostgresService{DB: db, Logger: zap.NewNop(), EventsTable: "events"}

	// Create an event
	eid, err := s.CreateEvent(
		"Test Event 1",
		"scoreboard-campaign",
		time.Now(),
		time.Now().Add(10*time.Minute),
		"guild-id",
		true,
	)
	require.NoError(err)
	assert.NotEmpty(eid)

	// Create another in the same guild
	_, err = s.CreateEvent(
		"Test Event 2",
		"scoreboard-campaign",
		time.Now(),
		time.Now().Add(10*time.Minute),
		"guild-id",
		true,
	)
	require.NoError(err)

	// Create another in a different guild
	_, err = s.CreateEvent(
		"Test Event 3",
		"scoreboard-campaign",
		time.Now(),
		time.Now().Add(10*time.Minute),
		"different-guild-id",
		true,
	)
	require.NoError(err)

	// Now these are all active, we should be able to do a couple things
	events, err := s.ListAllEvent()
	assert.NoError(err)
	assert.Equal(3, len(events))

	// List from one single guild
	events, err = s.ListEventsForGuild("guild-id")
	assert.NoError(err)
	assert.Equal(2, len(events))

	// List active from guild 1
	events, err = s.ListActiveEventsForGuild("guild-id")
	assert.NoError(err)
	assert.Equal(2, len(events))

	// Lets disable the first event
	err = s.SetEventStatus(eid, false)
	assert.NoError(err)

	// check it's indeed deactivated
	events, err = s.ListActiveEventsForGuild("guild-id")
	assert.NoError(err)
	assert.Equal(1, len(events))

	// lets set the end date of the event
	aBitLater := time.Now().Add(20 * time.Minute)
	err = s.SetEventEndDate(eid, aBitLater)
	assert.NoError(err)

	// check if the date is set right
	event, err := s.GetEvent(eid)
	assert.NoError(err)
	assert.Equal(aBitLater.Unix(), event.End.Unix())

	// lets update the name of the event
	err = s.UpdateEvent(
		eid,
		"New Event Name",
		event.EventType,
		event.Begin,
		event.End,
		event.GID,
		event.Active,
	)
	assert.NoError(err)

	// check if the name is set right
	event, err = s.GetEvent(eid)
	assert.NoError(err)
	assert.Equal("New Event Name", event.Name)

	// We delete an event
	err = s.DeleteEvent(eid)
	assert.NoError(err)

	// and check to make sure we only have one in guild 1
	events, err = s.ListEventsForGuild("guild-id")
	assert.NoError(err)
	assert.Equal(1, len(events))
}

func TestParticipation(t *testing.T) {
	db.MustExec(`
	CREATE TABLE events_test_par (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		guild_id text NOT NULL,
		name text NOT NULL,
		start_date timestamptz DEFAULT current_timestamp,
		end_date timestamptz NOT NULL,
		active boolean,
		event_type text
	);

	CREATE TABLE users_test_par (
		id text NOT NULL PRIMARY KEY,
		ign text NOT NULL
	);
	
	CREATE TABLE participation (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		user_id text,
		event_id uuid,
		participating boolean NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users_test_par(id) ON DELETE CASCADE,
		FOREIGN KEY (event_id) REFERENCES events_test_par(id) ON DELETE CASCADE,
		UNIQUE (user_id, event_id)
	);
	`)

	assert := assert.New(t)
	require := require.New(t)

	s := &meta.PostgresService{
		DB:                 db,
		Logger:             zap.NewNop(),
		EventsTable:        "events_test_par",
		IGNTable:           "users_test_par",
		ParticipationTable: "participation",
	}

	// make events and users
	eid1, err := s.CreateEvent(
		"Test Event 1",
		"scoreboard-campaign",
		time.Now(),
		time.Now().Add(10*time.Minute),
		"guild-id",
		true,
	)
	require.NoError(err)
	eid2, err := s.CreateEvent(
		"Test Event 2",
		"scoreboard-campaign",
		time.Now(),
		time.Now().Add(10*time.Minute),
		"guild-id",
		true,
	)
	require.NoError(err)

	err = s.CreateIGN("test-user-1", "test-ign-1")
	require.NoError(err)
	err = s.CreateIGN("test-user-2", "test-ign-2")
	require.NoError(err)

	// we make the users participate in the event
	pid, err := s.AddParticipation("test-user-1", eid1, true)
	assert.NoError(err)
	assert.NotEmpty(pid)

	_, err = s.AddParticipation("test-user-2", eid1, true)
	assert.NoError(err)

	// check and see that event 1 has two users with the right ign
	usersIn, _, err := s.ListUserForEvent(eid1)
	assert.NoError(err)
	assert.EqualValues(map[string]string{
		"test-user-1": "test-ign-1",
		"test-user-2": "test-ign-2",
	}, usersIn)

	// make sure no duplicates can be added
	_, err = s.AddParticipation("test-user-2", eid1, true)
	dupErr := &meta.ErrDuplicateEntry{}
	assert.ErrorAs(err, &dupErr)

	// but adding to another event is fine
	_, err = s.AddParticipation("test-user-2", eid2, true)
	assert.NoError(err)

	// make sure user 1 is in the event
	in, err := s.UserInEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.True(in)

	// get rid of user1 from the event by id
	err = s.DeleteParticipation(pid)
	assert.NoError(err)
	in, err = s.UserInEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.False(in)

	// try to remove this again should result in no record error
	err = s.DeleteParticipation(pid)
	noRecordErr := &meta.ErrNoRecord{}
	assert.ErrorAs(err, &noRecordErr)

	// we add user 1 back in
	pid, err = s.AddParticipation("test-user-1", eid1, true)
	assert.NoError(err)
	assert.NotEmpty(pid)

	in, err = s.UserInEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.True(in)

	// this time we remove user1 by userID and eventID
	err = s.DeleteParticipationByUserAndEvent("test-user-1", eid1)
	assert.NoError(err)
	in, err = s.UserInEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.False(in)

	_, _, err = s.GetParticipation("test-user-1", eid1)
	assert.ErrorAs(err, &noRecordErr)

	// we add user 1 back in
	pid, err = s.AddParticipation("test-user-1", eid1, true)
	assert.NoError(err)
	assert.NotEmpty(pid)

	// we update user 1's participation to be false by id
	err = s.SetParticipation(pid, false)
	assert.NoError(err)
	in, err = s.UserInEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.False(in)
	out, err := s.UserBailedEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.True(out)
	partID, stat, err := s.GetParticipation("test-user-1", eid1)
	assert.NoError(err)
	assert.False(stat)
	assert.NotEmpty(partID)
	// assert.Fail(partID)

	// we update user 1's participation by event id and user id
	err = s.SetParticipationByUserAndEvent("test-user-1", eid1, true)
	assert.NoError(err)
	in, err = s.UserInEvent("test-user-1", eid1)
	assert.NoError(err)
	assert.True(in)
}
