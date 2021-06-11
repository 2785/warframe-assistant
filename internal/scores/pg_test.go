package scores_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/2785/warframe-assistant/internal/scores"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
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

func TestNewWorkflow(t *testing.T) {
	eid1 := uuid.NewString()
	eid2 := uuid.NewString()

	pid1, pid2, pid3 := uuid.NewString(), uuid.NewString(), uuid.NewString()

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
	
	CREATE TABLE users (
		id text NOT NULL PRIMARY KEY,
		ign text NOT NULL
	);
	
	CREATE TABLE participation (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		user_id text,
		event_id uuid,
		participating boolean NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
		UNIQUE (user_id, event_id)
	);
	
	CREATE TABLE event_scores (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		score int NOT NULL,
		proof text NOT NULL,
		verified boolean DEFAULT FALSE,
		participation_id uuid,
		FOREIGN KEY (participation_id) REFERENCES participation(id) ON DELETE CASCADE
	);

	INSERT INTO users (
		id, ign
	) values (
		'test-user-1', 'test-ign-1'
	);

	INSERT INTO users (
		id, ign
	) values (
		'test-user-2', 'test-ign-2'
	);
	`)

	db.MustExec(fmt.Sprintf(`
	INSERT INTO events (
		id, guild_id, name, start_date, end_date, active, event_type
	) values (
		'%s', 'guild-1', 'test-event-1', CURRENT_TIMESTAMP, '2022-01-01 00:00:00', TRUE, 'scoreboard-campaign'
	);
	`, eid1))

	db.MustExec(fmt.Sprintf(`
	INSERT INTO events (
		id, guild_id, name, start_date, end_date, active, event_type
	) values (
		'%s', 'guild-1', 'test-event-2', CURRENT_TIMESTAMP, '2022-01-01 00:00:00', TRUE, 'scoreboard-campaign'
	);
	`, eid2))

	db.MustExec(fmt.Sprintf(`
	INSERT INTO participation (
		id, user_id, event_id, participating
	) values (
		'%s', 'test-user-1', '%s', TRUE
	);
	
	INSERT INTO participation (
		id, user_id, event_id, participating
	) values (
		'%s', 'test-user-2', '%s', TRUE
	);

	INSERT INTO participation (
		id, user_id, event_id, participating
	) values (
		'%s', 'test-user-1', '%s', TRUE
	);
	`, pid1, eid1, pid2, eid1, pid3, eid2))

	s := &scores.PostgresService{
		DB:                     db,
		ScoresTableName:        "event_scores",
		Logger:                 zap.NewNop(),
		UserIGNTableName:       "users",
		ParticipationTableName: "participation",
	}

	require := require.New(t)
	assert := assert.New(t)

	// make a new score claim
	sid1, err := s.ClaimScore(pid1, 3, "some-url")
	require.NoError(err)
	assert.NotEmpty(sid1)

	// make sure the verification status is as expected - 1 total, 0 verified
	total, verified, err := s.VerificationStatus(eid1)
	assert.NoError(err)
	assert.Equal(1, total)
	assert.Equal(0, verified)

	// make sure we can get a record without specifying eid
	record, err := s.GetOneUnverified()
	assert.NoError(err)
	assert.Equal("test-user-1", record.UID)

	// make sure we can get a record while specifying eid
	record, err = s.GetOneUnverifiedForEvent(eid1)
	assert.NoError(err)
	assert.Equal("test-ign-1", record.IGN)

	// event 2 should not have any record
	_, err = s.GetOneUnverifiedForEvent(eid2)
	assert.Error(err)
	nr := &scores.ErrNoRecord{}
	assert.ErrorAs(err, &nr)

	// verify the submission
	err = s.Verify(sid1)
	assert.NoError(err)

	// make sure the verification status is as expected - 1 total, 1 verified
	total, verified, err = s.VerificationStatus(eid1)
	assert.NoError(err)
	assert.Equal(1, total)
	assert.Equal(1, verified)

	_, err = s.GetOneUnverifiedForEvent(eid1)
	assert.Error(err)
	assert.ErrorAs(err, &nr)

	leaderboard, err := s.MakeReportScoreSum(eid1)
	assert.NoError(err)
	assert.Equal([]scores.SummaryRecord{
		{
			UID:   "test-user-1",
			IGN:   "test-ign-1",
			Score: 3,
		},
	}, leaderboard)

	// make another submission user 2 to take over user 1
	_, err = s.ClaimScore(pid2, 5, "some-url")
	require.NoError(err)

	// lets verify it
	score2, err := s.GetOneUnverifiedForEvent(eid1)
	require.NoError(err)
	err = s.Verify(score2.ID)
	require.NoError(err)

	// and check verification status / leaderboard
	total, verified, err = s.VerificationStatus(eid1)
	assert.NoError(err)
	assert.Equal(2, total)
	assert.Equal(2, verified)

	leaderboard, err = s.MakeReportScoreSum(eid1)
	assert.NoError(err)
	assert.Equal([]scores.SummaryRecord{
		{
			UID:   "test-user-2",
			IGN:   "test-ign-2",
			Score: 5,
		},
		{
			UID:   "test-user-1",
			IGN:   "test-ign-1",
			Score: 3,
		},
	}, leaderboard)

	// make another submission by user 1 with 1 score
	sid3, err := s.ClaimScore(pid1, 1, "http://google.ca")
	require.NoError(err)

	// check leaderboard now
	leaderboard, err = s.MakeReportScoreSum(eid1)
	assert.NoError(err)
	assert.Equal([]scores.SummaryRecord{
		{
			UID:   "test-user-2",
			IGN:   "test-ign-2",
			Score: 5,
		},
		{
			UID:   "test-user-1",
			IGN:   "test-ign-1",
			Score: 3,
		},
	}, leaderboard)

	// lets buff this score up
	err = s.UpdateScoreAndVerify(sid3, 9000)
	assert.NoError(err)

	// check the leaderboard now
	leaderboard, err = s.MakeReportScoreSum(eid1)
	assert.NoError(err)
	assert.Equal([]scores.SummaryRecord{
		{
			UID:   "test-user-1",
			IGN:   "test-ign-1",
			Score: 9003,
		},
		{
			UID:   "test-user-2",
			IGN:   "test-ign-2",
			Score: 5,
		},
	}, leaderboard)

	// check the leaderboard in top score mode now
	leaderboard, err = s.MakeReportScoreTop(eid1)
	assert.NoError(err)
	assert.Equal([]scores.SummaryRecord{
		{
			UID:   "test-user-1",
			IGN:   "test-ign-1",
			Score: 9000,
		},
		{
			UID:   "test-user-2",
			IGN:   "test-ign-2",
			Score: 5,
		},
	}, leaderboard)

	// buffed too much, lets remove it
	err = s.DeleteScore(sid3)
	assert.NoError(err)

	// check leaderboard now
	leaderboard, err = s.MakeReportScoreSum(eid1)
	assert.NoError(err)
	assert.Equal([]scores.SummaryRecord{
		{
			UID:   "test-user-2",
			IGN:   "test-ign-2",
			Score: 5,
		},
		{
			UID:   "test-user-1",
			IGN:   "test-ign-1",
			Score: 3,
		},
	}, leaderboard)
}
