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

// // can't be fucked to unit test everything carefully for now
// func TestWorkflow(t *testing.T) {
// 	// make the table
// 	db.MustExec(`CREATE TABLE devtest (
// 		submissionID uuid DEFAULT gen_random_uuid (),
// 		userID text NOT NULL,
// 		ign text NOT NULL,
// 		score int NOT NULL,
// 		proof text NOT NULL,
// 		verified boolean DEFAULT FALSE
// 	);`)

// 	s := &scores.PostgresService{
// 		DB:     db,
// 		Table:  "devtest",
// 		Logger: zap.NewNop(),
// 	}

// 	// add some records
// 	id, err := s.AddUnverified("test-1", "test-1", 1, "some-url")
// 	require.NoError(t, err)
// 	assert.NotEmpty(t, id)
// 	_, err = s.AddUnverified("test-1", "test-1", 1, "some-other-url")
// 	assert.NoError(t, err)
// 	_, err = s.AddUnverified("test-2", "test-2", 1, "some-other-url")
// 	assert.NoError(t, err)

// 	// make sure records are as expected for now
// 	total, verified, pending, err := s.VerificationStatus()
// 	require.NoError(t, err)
// 	assert.Equal(t, 3, total)
// 	assert.Equal(t, 3, pending)
// 	assert.Equal(t, 0, verified)

// 	// grab one
// 	id, _, _, _, _, err = s.GetOneUnverified()
// 	require.NoError(t, err)

// 	// verify one
// 	err = s.Verify(id)
// 	require.NoError(t, err)

// 	// make sure records are as expected for now
// 	total, verified, pending, err = s.VerificationStatus()
// 	require.NoError(t, err)
// 	assert.Equal(t, 3, total)
// 	assert.Equal(t, 2, pending)
// 	assert.Equal(t, 1, verified)

// 	// verify another
// 	id, _, _, _, _, err = s.GetOneUnverified()
// 	require.NoError(t, err)
// 	err = s.Verify(id)
// 	require.NoError(t, err)

// 	// and another, this time we update the score to 5
// 	id, _, _, _, _, err = s.GetOneUnverified()
// 	require.NoError(t, err)
// 	err = s.UpdateScore(id, 5)
// 	require.NoError(t, err)

// 	// should error now with no record err if we try to get again
// 	_, _, _, _, _, err = s.GetOneUnverified()
// 	errNoRecord := &scores.ErrNoRecord{}
// 	assert.ErrorAs(t, err, &errNoRecord)

// 	// take a look at verification status again
// 	total, verified, pending, err = s.VerificationStatus()
// 	require.NoError(t, err)
// 	assert.Equal(t, 3, total)
// 	assert.Equal(t, 0, pending)
// 	assert.Equal(t, 3, verified)

// 	// now we look at scoreboard
// 	scoreboard, err := s.ScoreReport()
// 	require.NoError(t, err)
// 	assert.Equal(t, []scores.SummaryRecord{
// 		{
// 			IGN:    "test-2",
// 			UserID: "test-2",
// 			Sum:    5,
// 		},
// 		{
// 			IGN:    "test-1",
// 			UserID: "test-1",
// 			Sum:    2,
// 		},
// 	}, scoreboard)

// 	// try deleting one for shits and giggles
// 	err = s.DeleteRecord(id)
// 	require.NoError(t, err)

// 	// now we look at scoreboard again to make sure the test-2 user is gone
// 	scoreboard, err = s.ScoreReport()
// 	require.NoError(t, err)
// 	assert.Equal(t, []scores.SummaryRecord{
// 		{
// 			IGN:    "test-1",
// 			UserID: "test-1",
// 			Sum:    2,
// 		},
// 	}, scoreboard)
// }

func TestNewWorkflow(t *testing.T) {
	event_id := uuid.NewString()

	db.MustExec(`
	CREATE TABLE events (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		guild_id text NOT NULL,
		name text NOT NULL,
		start_date timestamptz DEFAULT current_timestamp,
		end_date timestamptz NOT NULL,
		active boolean
	);

	CREATE TABLE users (
		id text NOT NULL PRIMARY KEY,
		ign text NOT NULL
	);
	
	CREATE TABLE event_scores (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		score int NOT NULL,
		proof text NOT NULL,
		verified boolean DEFAULT FALSE,
		user_id text,
		event_id uuid,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
	);
	
	CREATE TABLE participation (
		id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
		user_id text,
		event_id uuid,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
	);

	INSERT INTO users (
		id, ign
	) values (
		'test-user-1', 'test-user-1'
	);

	INSERT INTO users (
		id, ign
	) values (
		'test-user-2', 'test-user-2'
	);
	`)

	db.MustExec(fmt.Sprintf(`
	INSERT INTO events (
		id, guild_id, name, start_date, end_date, active
	) values (
		'%s', 'guild-id', 'test-event-1', CURRENT_TIMESTAMP, '2022-01-01 00:00:00', TRUE
	)
	`, event_id))

	s := &scores.PostgresService{
		DB:               db,
		ScoresTableName:  "event_scores",
		Logger:           zap.NewNop(),
		UserIGNTableName: "users",
	}

	user1SubmissionID, err := s.SubmitScore("test-user-1", event_id, "some-url", 3)
	require.NoError(t, err)
	_, err = s.SubmitScore("test-user-2", event_id, "some-url", 2)
	require.NoError(t, err)
	_, err = s.SubmitScore("test-user-2", event_id, "some-url", 4)
	require.NoError(t, err)

	total, verified, pending, err := s.VerificationStatusForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 3, pending)
	assert.Equal(t, 0, verified)

	// grab one
	entry, err := s.GetOneUnverifiedSubmission(event_id)
	require.NoError(t, err)

	// verify one
	err = s.VerifySubmission(entry.ID)
	require.NoError(t, err)

	// make sure records are as expected for now
	total, verified, pending, err = s.VerificationStatusForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 2, pending)
	assert.Equal(t, 1, verified)

	// verify another
	entry, err = s.GetOneUnverifiedSubmission(event_id)
	require.NoError(t, err)
	err = s.VerifySubmission(entry.ID)
	require.NoError(t, err)

	// lets pull the score board, now we should have user 1 ahead of user 2
	leaderboard, err := s.ScoreReportForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, []scores.ScoreSummary{
		{
			UID:        "test-user-1",
			IGN:        "test-user-1",
			TotalScore: 3,
		}, {
			UID:        "test-user-2",
			IGN:        "test-user-2",
			TotalScore: 2,
		},
	}, leaderboard)

	// and another
	entry, err = s.GetOneUnverifiedSubmission(event_id)
	require.NoError(t, err)
	err = s.VerifySubmission(entry.ID)
	require.NoError(t, err)

	// lets pull the score board again, now we should have user 2 ahead of user 1
	leaderboard, err = s.ScoreReportForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, []scores.ScoreSummary{

		{
			UID:        "test-user-2",
			IGN:        "test-user-2",
			TotalScore: 6,
		},
		{
			UID:        "test-user-1",
			IGN:        "test-user-1",
			TotalScore: 3,
		},
	}, leaderboard)

	// should error now with no record err if we try to get again
	_, err = s.GetOneUnverifiedSubmission(event_id)
	errNoRecord := &scores.ErrNoRecord{}
	assert.ErrorAs(t, err, &errNoRecord)

	// lets buff user 1 up :)
	err = s.UpdateSubmissionScore(user1SubmissionID, 9000)
	require.NoError(t, err)

	// lets pull the score board, now we should have user 1 ahead of user 2
	leaderboard, err = s.ScoreReportForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, []scores.ScoreSummary{
		{
			UID:        "test-user-1",
			IGN:        "test-user-1",
			TotalScore: 9000,
		},
		{
			UID:        "test-user-2",
			IGN:        "test-user-2",
			TotalScore: 6,
		},
	}, leaderboard)

	// take a look at verification status again
	total, verified, pending, err = s.VerificationStatusForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 0, pending)
	assert.Equal(t, 3, verified)

	// buffed user 1 too much, lets now nerf it to the ground
	err = s.DeleteSubmission(user1SubmissionID)
	require.NoError(t, err)

	// and we should see user 1 gone
	leaderboard, err = s.ScoreReportForEvent(event_id)
	require.NoError(t, err)
	assert.Equal(t, []scores.ScoreSummary{
		{
			UID:        "test-user-2",
			IGN:        "test-user-2",
			TotalScore: 6,
		},
	}, leaderboard)
}
