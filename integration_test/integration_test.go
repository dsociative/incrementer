// +build integration

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/dsociative/incrementer/api"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	client               api.Incrementer
	pool                 *dockertest.Pool
	redisContainer       *dockertest.Resource
	incrementerContainer *dockertest.Resource
)

func initPool() *dockertest.Pool {
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	return p
}

func initRedisContainer() {
	resource, err := pool.Run("redis", "latest", nil)
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	redisContainer = resource
}

func initIncrementerContainer() {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "incrementer",
		Tag:        "latest",
		Links:      []string{redisContainer.Container.Name + ":redis"},
		Cmd:        []string{"--redis", "redis:6379"},
	})
	if err != nil {
		log.Fatalf("Could not start incrementer: %s", err)
	}
	client = api.NewIncrementerProtobufClient(
		"http://localhost:"+resource.GetPort("8080/tcp"),
		&http.Client{},
	)
	if err = pool.Retry(func() error {
		_, err := client.GetNumber(context.Background(), &api.Empty{})
		return err
	}); err != nil {
		log.Fatalf("Could not connect to incrementer: %s", err)
	}
	incrementerContainer = resource
}

func TestMain(m *testing.M) {
	pool = initPool()
	os.Exit(m.Run())
}

func TestIncrementer(t *testing.T) {
	for _, f := range []func(*testing.T){
		getNumber,
		incrementNumber,
		setSettings,
		restartIncrementer,
		restartDBWithLossAllData,
	} {
		initRedisContainer()
		initIncrementerContainer()
		t.Run(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), f)
		assert.NoError(t, redisContainer.Close())
		assert.NoError(t, incrementerContainer.Close())
	}
}

func getNumber(t *testing.T) {
	n, err := client.GetNumber(context.Background(), &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n.Number)
}

func incrementNumber(t *testing.T) {
	for i := 1; i < 10; i++ {
		n, err := client.IncrementNumber(context.Background(), &api.Empty{})
		require.NoError(t, err)
		assert.Equal(t, int64(i), n.Number)

		n, err = client.GetNumber(context.Background(), &api.Empty{})
		assert.NoError(t, err)
		assert.Equal(t, int64(i), n.Number)
	}
}

func setSettings(t *testing.T) {
	n, err := client.IncrementNumber(context.Background(), &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n.Number)

	_, err = client.SetSettings(context.Background(), &api.Setting{
		Maximum: 20,
		Step:    5,
	})
	assert.NoError(t, err)

	for _, expectedNum := range []int64{6, 11, 16, 0} {
		n, err := client.IncrementNumber(context.Background(), &api.Empty{})
		assert.NoError(t, err)
		assert.Equal(t, expectedNum, n.Number)
	}
}

func restartIncrementer(t *testing.T) {
	n, err := client.IncrementNumber(context.Background(), &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n.Number)

	// Kill incrementer container
	assert.NoError(t, incrementerContainer.Close())

	// Check command error
	n, err = client.GetNumber(context.Background(), &api.Empty{})
	assert.Error(t, err)
	assert.Nil(t, n)

	// Start new incrementer container
	initIncrementerContainer()

	// Check data not losed
	n, err = client.GetNumber(context.Background(), &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n.Number)
}

func restartDBWithLossAllData(t *testing.T) {
	n, err := client.IncrementNumber(context.Background(), &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n.Number)

	// Kill redis container
	assert.NoError(t, redisContainer.Close())

	// Check provision error
	n, err = client.IncrementNumber(context.Background(), &api.Empty{})
	assert.Error(t, err)
	assert.Nil(t, n)

	// Start new redis container and check command error without initial settings
	initRedisContainer()
	n, err = client.IncrementNumber(context.Background(), &api.Empty{})
	assert.Error(t, err)

	// Set initial setting
	_, err = client.SetSettings(context.Background(), &api.Setting{
		Maximum: 6,
		Step:    3,
	})
	assert.NoError(t, err)

	// Check
	n, err = client.IncrementNumber(context.Background(), &api.Empty{})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), n.Number)
}
