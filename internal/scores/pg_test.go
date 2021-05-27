package scores_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/2785/warframe-assistant/internal/scores"
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

// can't be fucked to unit test everything carefully for now
func TestWorkflow(t *testing.T) {
	// make the table
	db.MustExec(`CREATE TABLE devtest (
		submissionID uuid DEFAULT gen_random_uuid (),
		userID text NOT NULL,
		ign text NOT NULL,
		score int NOT NULL,
		proof text NOT NULL,
		verified boolean DEFAULT FALSE
	);`)

	s := &scores.PostgresService{
		DB:     db,
		Table:  "devtest",
		Logger: zap.NewNop(),
	}

	// add some records
	id, err := s.AddUnverified("test-1", "test-1", 1, "some-url")
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	_, err = s.AddUnverified("test-1", "test-1", 1, "some-other-url")
	assert.NoError(t, err)
	_, err = s.AddUnverified("test-2", "test-2", 1, "some-other-url")
	assert.NoError(t, err)

	// make sure records are as expected for now
	total, verified, pending, err := s.VerificationStatus()
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 3, pending)
	assert.Equal(t, 0, verified)

	// grab one
	id, _, _, _, _, err = s.GetOneUnverified()
	require.NoError(t, err)

	// verify one
	err = s.Verify(id)
	require.NoError(t, err)

	// make sure records are as expected for now
	total, verified, pending, err = s.VerificationStatus()
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 2, pending)
	assert.Equal(t, 1, verified)

	// verify another
	id, _, _, _, _, err = s.GetOneUnverified()
	require.NoError(t, err)
	err = s.Verify(id)
	require.NoError(t, err)

	// and another, this time we update the score to 5
	id, _, _, _, _, err = s.GetOneUnverified()
	require.NoError(t, err)
	err = s.UpdateScore(id, 5)
	require.NoError(t, err)

	// should error now with no record err if we try to get again
	_, _, _, _, _, err = s.GetOneUnverified()
	errNoRecord := &scores.ErrNoRecord{}
	assert.ErrorAs(t, err, &errNoRecord)

	// take a look at verification status again
	total, verified, pending, err = s.VerificationStatus()
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Equal(t, 0, pending)
	assert.Equal(t, 3, verified)

	// now we look at scoreboard
	scoreboard, err := s.ScoreReport()
	require.NoError(t, err)
	assert.Equal(t, []scores.SummaryRecord{
		{
			IGN:    "test-2",
			UserID: "test-2",
			Sum:    5,
		},
		{
			IGN:    "test-1",
			UserID: "test-1",
			Sum:    2,
		},
	}, scoreboard)

	// try deleting one for shits and giggles
	err = s.DeleteRecord(id)
	require.NoError(t, err)

	// now we look at scoreboard again to make sure the test-2 user is gone
	scoreboard, err = s.ScoreReport()
	require.NoError(t, err)
	assert.Equal(t, []scores.SummaryRecord{
		{
			IGN:    "test-1",
			UserID: "test-1",
			Sum:    2,
		},
	}, scoreboard)
}
