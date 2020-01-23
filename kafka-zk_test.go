package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestKafkaZookeeper(t *testing.T) {
	ctx := context.Background()

	/*
		docker run -ti --rm --name kafka-zk \
			-e ADVERTISED_HOST=localhost \
			-e ADVERTISED_PORT=9092 \
			-p 2181:2181 \
			-p 9092:9092 \
			spotify/kafka
	*/

	t.Log("Starting Kafka+Zookeeper container...")
	contReq := testcontainers.ContainerRequest{
		Name:  "kafka-zk",
		Image: "spotify/kafka",
		Env: map[string]string{
			"ADVERTISED_HOST": "localhost",
			"ADVERTISED_PORT": "9092",
		},
		ExposedPorts: []string{
			"2181:2181",
			"9092:9092",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("spawned: 'kafka' with pid").WithPollInterval(1 * time.Second),
		),
	}
	cont, startErr := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: contReq,
			Started:          true,
		},
	)
	if startErr != nil {
		t.Error(startErr)
	}
	defer cont.Terminate(ctx)

	host, hostErr := cont.Host(ctx)
	if hostErr != nil {
		t.Error(hostErr)
	}
	zkPort, zkPortErr := cont.MappedPort(ctx, "2181/tcp")
	if zkPortErr != nil {
		t.Error(zkPortErr)
	}
	kfPort, kfPortErr := cont.MappedPort(ctx, "9092/tcp")
	if kfPortErr != nil {
		t.Error(kfPortErr)
	}
	id := cont.GetContainerID()
	t.Logf("Kafka+Zookeeper started: host %s, zk-port %s, kf-port %s, id %s\n", host, zkPort, kfPort, id)
	// t.Log("Zookeeper started with id", id)

	expPorts, err := cont.Ports(ctx)
	if err != nil {
		t.Error(err)
	}
	t.Log("Kafka+Zookeeper exposed ports:")
	for k, v := range expPorts {
		t.Logf("\t %s -> %v \n", k, v)
	}

	assert.NotEqual(t, "", host)
	assert.NotEqual(t, "", zkPort)
	assert.NotEqual(t, "", kfPort)
	assert.NotEqual(t, "", id)
}
