package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedis(t *testing.T) {
	ctx := context.Background()

	fmt.Println("Starting Redis container...")
	req := testcontainers.ContainerRequest{
		Image:        "redis",
		ExposedPorts: []string{"6379/tcp"},
		// WaitingFor:   wait.ForHTTP("/"),
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections").WithPollInterval(200 * time.Millisecond),
		),
	}
	cont, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer cont.Terminate(ctx)

	ip, err := cont.Host(ctx)
	if err != nil {
		t.Error(err)
	}
	port, err := cont.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("Redis started: host %s, port %s\n", ip, port)

	fmt.Println("Testing use cases...")
	assert.NotEqual(t, "", ip)
	assert.NotEqual(t, "", port)
}
