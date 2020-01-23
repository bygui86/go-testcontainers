package main

import (
	"context"
	"testing"
	"time"

	// "github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// WARN: not working, no idea why!
func TestKafka(t *testing.T) {
	ctx := context.Background()

	zkCont, network := startZookeeper(t, ctx)
	defer zkCont.Terminate(ctx)
	defer network.Remove(ctx)

	kfCont := startKafka(t, ctx)
	defer kfCont.Terminate(ctx)

	assert.True(t, true)
}

func startZookeeper(t *testing.T, ctx context.Context) (testcontainers.Container, testcontainers.Network) {
	/* host network
	docker run -ti --rm --name zookeeper \
		-e ZOOKEEPER_CLIENT_PORT=2181 \
		-e ZOOKEEPER_TICK_TIME=2000 \
		-p 2181:2181 \
		--net host \
		confluentinc/cp-zookeeper:5.2.1
	*/
	/* other networks
	docker run -ti --rm --name zookeeper \
		-e ZOOKEEPER_CLIENT_PORT=2181 \
		-e ZOOKEEPER_TICK_TIME=2000 \
		-p 2181:2181 \
		--net testing \
		confluentinc/cp-zookeeper:5.2.1
	*/

	t.Log("Starting Zookeeper container...")
	contReq := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:  "zookeeper",
			Image: "confluentinc/cp-zookeeper:5.2.1",
			Env: map[string]string{
				"ZOOKEEPER_CLIENT_PORT": "2181",
				"ZOOKEEPER_TICK_TIME":   "2000",
			},
			ExposedPorts: []string{"2181/tcp"},
			Networks:     []string{"testing"},
			NetworkAliases: map[string][]string{
				"testing": {"zookeeper"},
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("binding to port 0.0.0.0/0.0.0.0:2181").WithPollInterval(500 * time.Millisecond),
			),
		},
		// Started:          true,
	}

	provider, provErr := contReq.ProviderType.GetProvider()
	if provErr != nil {
		t.Error(provErr)
	}
	network := createNetwork(t, ctx, provider)

	cont, startErr := testcontainers.GenericContainer(ctx, contReq)
	if startErr != nil {
		t.Error(startErr)
	}

	// host, hostErr := cont.Host(ctx)
	// if hostErr != nil {
	// 	t.Error(hostErr)
	// }
	// port, portErr := cont.MappedPort(ctx, "2181/tcp")
	// if portErr != nil {
	// 	t.Error(portErr)
	// }
	id := cont.GetContainerID()
	// t.Logf("Zookeeper started: host %s, port %s, id %s\n", host, port, id)
	t.Log("Zookeeper started with id", id)

	// expPorts, err := cont.Ports(ctx)
	// if err != nil {
	// 	t.Error(err)
	// }
	// t.Log("Zookeeper exposed ports:")
	// for k, v := range expPorts {
	// 	t.Logf("\t %s -> %v \n", k, v)
	// }

	// assert.NotEqual(t, "", host)
	// assert.NotEqual(t, "", port)
	assert.NotEqual(t, "", id)

	return cont, network
}

func createNetwork(t *testing.T, ctx context.Context, provider testcontainers.GenericProvider) testcontainers.Network {
	/*
		docker network create testing
	*/

	t.Log("Creating testing network...")
	network, _ := provider.CreateNetwork(ctx, testcontainers.NetworkRequest{
		Driver:         "",
		CheckDuplicate: true,
		Internal:       false,
		Name:           "testing",
		Attachable:     true,
	})

	return network
}

func startKafka(t *testing.T, ctx context.Context) testcontainers.Container {
	/* host network
	docker run -ti --rm --name kafka \
		-e KAFKA_BROKER_ID=1 \
		-e KAFKA_ZOOKEEPER_CONNECT=localhost:2181 \
		-e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT \
		-e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT_HOST://localhost:9092,PLAINTEXT://localhost:29092 \
		-e AUTO_LEADER_REBALANCE_ENABLE=FALSE \
		-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
		-e KAFKA_JMX_PORT=49999 \
		-e KAFKA_JMX_HOSTNAME=localhost \
		-p 9092:9092 \
		-p 29092:29092 \
		-p 49999:49999 \
		--net host \
		confluentinc/cp-enterprise-kafka:5.3.1
	*/
	/* other networks
	docker run -ti --rm --name kafka \
		-e KAFKA_BROKER_ID=1 \
		-e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
		-e KAFKA_LISTENER_SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT \
		-e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT_HOST://kafka:9092,PLAINTEXT://kafka:29092 \
		-e AUTO_LEADER_REBALANCE_ENABLE=FALSE \
		-e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
		-e KAFKA_JMX_PORT=49999 \
		-e KAFKA_JMX_HOSTNAME=kafka \
		-p 9092:9092 \
		-p 29092:29092 \
		-p 49999:49999 \
		--net testing \
		confluentinc/cp-enterprise-kafka:5.3.1
	*/

	t.Log("Starting Kafka container...")
	contReq := testcontainers.ContainerRequest{
		Name:  "kafka",
		Image: "confluentinc/cp-enterprise-kafka:5.3.1",
		Env: map[string]string{
			// host network
			// "KAFKA_BROKER_ID": "1",
			// "KAFKA_ZOOKEEPER_CONNECT":                "localhost:2181",
			// "KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			// "KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT_HOST://localhost:9092,PLAINTEXT://localhost:29092",
			// "AUTO_LEADER_REBALANCE_ENABLE":           "FALSE",
			// "KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			// "KAFKA_JMX_PORT":                         "49999",
			// "KAFKA_JMX_HOSTNAME":                     "localhost",
			// other networks
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:2181",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT_HOST://kafka:9092,PLAINTEXT://kafka:29092",
			"AUTO_LEADER_REBALANCE_ENABLE":           "FALSE",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_JMX_PORT":                         "49999",
			"KAFKA_JMX_HOSTNAME":                     "kafka",
		},
		ExposedPorts: []string{
			"9092/tcp",
			"29092/tcp",
			"49999/tcp",
		},
		Networks: []string{"testing"},
		NetworkAliases: map[string][]string{
			"testing": {"kafka"},
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("[KafkaServer id=1] started").WithPollInterval(1 * time.Second),
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

	// host, hostErr := cont.Host(ctx)
	// if hostErr != nil {
	// 	t.Error(hostErr)
	// }
	// port, portErr := cont.MappedPort(ctx, "9092/tcp")
	// if portErr != nil {
	// 	t.Error(portErr)
	// }
	id := cont.GetContainerID()
	// t.Logf("Kafka started: host %s, port %s, id %s\n", host, port, id)
	t.Log("Kafka started with id", id)

	// assert.NotEqual(t, "", host)
	// assert.NotEqual(t, "", port)
	assert.NotEqual(t, "", id)

	return cont
}
