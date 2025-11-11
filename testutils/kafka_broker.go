//coverage:ignore file

package testutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type kafkaContainer struct {
	zookeeper StoppableContainer
	kafka     StoppableContainer
	network   testcontainers.Network
}

func NewKafkaContainer() StoppableContainer {
	KafkaNetworkName := fmt.Sprintf("kafka-network-%d", time.Now().UnixNano())
	kafkaNetwork := createNetwork(KafkaNetworkName)

	return &kafkaContainer{
		network: kafkaNetwork,
		zookeeper: &standardContainer{
			req: testcontainers.ContainerRequest{
				Image:        "bitnami/zookeeper:3",
				ExposedPorts: []string{"2181/tcp"},
				WaitingFor:   wait.ForAll(wait.ForListeningPort("2181/tcp"), wait.ForLog("binding to port")),
				Env: map[string]string{
					"ALLOW_ANONYMOUS_LOGIN": "YES",
				},
				Networks: []string{KafkaNetworkName},
				Name:     "zookeeper",
			},
		},

		kafka: &standardContainer{
			req: testcontainers.ContainerRequest{
				Image:        "bitnami/kafka:3",
				ExposedPorts: []string{"9093/tcp"},
				WaitingFor:   wait.ForAll(wait.ForListeningPort("9093/tcp"), wait.ForLog("started (kafka.server.KafkaServer)")),
				Env: map[string]string{
					"KAFKA_CFG_LISTENERS":                      "PLAINTEXT://0.0.0.0:9093,BROKER://0.0.0.0:9092",
					"KAFKA_CFG_ADVERTISED_LISTENERS":           "BROKER://localhost:9092",
					"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "BROKER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
					"KAFKA_CFG_INTER_BROKER_LISTENER_NAME":     "BROKER",
					"KAFKA_CFG_BROKER_ID":                      "1",
					"ALLOW_PLAINTEXT_LISTENER":                 "yes",
					"KAFKA_CFG_ZOOKEEPER_CONNECT":              "zookeeper:2181",
					"KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE":      "true",
					"BITNAMI_DEBUG":                            "true",
				},
				Cmd:      []string{},
				Networks: []string{KafkaNetworkName},
			},
		},
	}
}

func (k *kafkaContainer) Start() {
	k.zookeeper.Start()
	k.kafka.Start()
	status, _, err := k.kafka.Execute([]string{
		"kafka-configs.sh",
		"--alter",
		"--bootstrap-server", "BROKER://localhost:9092",
		"--entity-type", "brokers",
		"--entity-name", "1",
		"--add-config",
		"advertised.listeners=[" + strings.Join([]string{fmt.Sprintf("BROKER://%s:%s", k.kafka.GetHost(), "9092"), fmt.Sprintf("PLAINTEXT://%s:%s", k.kafka.GetHost(), k.kafka.GetPort())}, ",") + "]",
	})
	if err != nil {
		panic(err)
	} else if status != 0 {
		panic(errors.New("unable to configure kafka"))
	}
}

func (k *kafkaContainer) Stop() {
	k.kafka.Stop()
	k.zookeeper.Stop()
	if err := k.network.Remove(context.TODO()); err != nil {
		log.Error(err)
	}
}

func (k *kafkaContainer) Execute(_ []string) (int, io.Reader, error) {
	panic("not implemented")
}

func (k *kafkaContainer) GetHost() string {
	return k.kafka.GetHost()
}

func (k *kafkaContainer) GetPort() string {
	return k.kafka.GetPort()
}

func (k *kafkaContainer) GetZkHost() string {
	return k.zookeeper.GetHost()
}

func (k *kafkaContainer) GetZkPort() string {
	return k.zookeeper.GetPort()
}

func (k *kafkaContainer) GetBrokerAddress() string {
	return fmt.Sprintf("%s:%s", k.kafka.GetHost(), k.kafka.GetPort())
}
