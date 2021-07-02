package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var redisDsn string

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost == "" {
		dockerHost = "localhost"
	}

	redisContainer, err := pool.Run("redis", "6.2.3-alpine", []string{})

	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	redisDsn = fmt.Sprintf("redis://%s:%s", dockerHost, redisContainer.GetPort("6379/tcp"))
	rOpt, err := redis.ParseURL(redisDsn)

	if err != nil {
		log.Printf("Error parsing DSN: %s", err)
		if err := pool.Purge(redisContainer); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
		os.Exit(1)
	}

	rClient := redis.NewClient(rOpt)

	if err := pool.Retry(func() error {
		var err error
		ctx := context.Background()
		_, err = rClient.Ping(ctx).Result()
		if err != nil {
			fmt.Printf("Ping error: %s", err)
		}
		return err
	}); err != nil {
		log.Fatalf("Could not connect to redis docker container: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(redisContainer); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestRedisCache(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	rCache, err := NewRedis(redisDsn, 2*time.Minute)
	require.NoError(err)
	assert.NotEmpty(rCache)

	type thing struct {
		F1 string
		F2 int
	}

	thing1 := &thing{
		F1: "thing1",
		F2: 5,
	}

	err = rCache.Set("thing1", thing1)
	assert.NoError(err)

	wantThing1 := &thing{}
	err = rCache.Get("thing1", wantThing1)
	assert.NoError(err)
	assert.Equal(thing1, wantThing1)

	err = rCache.Drop("thing1")
	assert.NoError(err)

	wantThing1 = &thing{}
	err = rCache.Get("thing1", wantThing1)
	assert.Error(err)
	assert.True(AsErrNoRecord(err))

	wantThing1 = &thing{}
	err = rCache.Once("thing1", wantThing1, func() (interface{}, error) {
		return thing1, nil
	})
	assert.NoError(err)
	assert.Equal(thing1, wantThing1)

	wantThing1 = &thing{}
	err = rCache.Get("thing1", wantThing1)
	assert.NoError(err)
	assert.Equal(thing1, wantThing1)
}
