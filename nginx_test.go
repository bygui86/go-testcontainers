package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNginx(t *testing.T) {
	ctx := context.Background()

	fmt.Println("Starting NGINX container...")
	req := testcontainers.ContainerRequest{
		Image:        "nginx",
		ExposedPorts: []string{"80/tcp"},
		// Networks:     []string{"host"},
		Networks:   []string{"testing"},
		WaitingFor: wait.ForHTTP("/"),
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
	port, err := cont.MappedPort(ctx, "80/tcp")
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("NGINX started: host %s, port %s\n", ip, port)

	expPorts, err := cont.Ports(ctx)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("NGINX exposed ports:")
	for k, v := range expPorts {
		fmt.Printf("\t %s -> %v \n", k, v)
	}

	fmt.Println("Testing use cases...")
	// port
	assert.NotEqual(t, "", ip)
	assert.NotEqual(t, "", port)
	// response
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// if resp.StatusCode != http.StatusOK {
	// 	t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	// }
}
