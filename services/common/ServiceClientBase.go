package common

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/TAULargeScaleWorkshop/DTOY/config"
	RegistryServiceClient "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/client"
	"github.com/pebbe/zmq4"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type ServiceClientBase[client_t any] struct {
	RegistryAddresses []string
	ServiceName       string
	CreateClient      func(grpc.ClientConnInterface) client_t
}

func (obj *ServiceClientBase[client_t]) LoadRegistryAddresses() {
	configFile := "./configurations/RegistryAddresses.yaml"
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error reading registry yaml file: %v", err)
		os.Exit(2)
	}
	var config config.RegistryServiceConfig
	err = yaml.Unmarshal(configData, &config) // parses YAML
	if err != nil {
		log.Fatalf("error unmarshaling registry addresses data: %v", err)
		os.Exit(3)
	}
	if len(config.RegistryAddresses) <= 0 {
		log.Fatalf("registry addresses yaml file does not include any enteries")
		os.Exit(4)
	}
	obj.RegistryAddresses = config.RegistryAddresses
}

func (obj *ServiceClientBase[client_t]) Connect() (res client_t, closeFunc func(), err error) {

	registry_client := RegistryServiceClient.NewRegistryServiceClient(obj.RegistryAddresses)
	if registry_client == nil {
		var empty client_t
		return empty, nil, fmt.Errorf("failed to create registry client. no registry addresses were provided")
	}

	services, err := registry_client.Discover(obj.ServiceName)

	if err != nil {
		var empty client_t
		return empty, nil, fmt.Errorf("failed to connect client to registry: %v", err)
	}

	serviceAddress, err := PickNode(services)

	if err != nil {
		var empty client_t
		return empty, nil, fmt.Errorf("no services available for service type: %v", obj.ServiceName)
	}

	conn, err := grpc.Dial(serviceAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		var empty client_t
		return empty, nil, fmt.Errorf("failed to connect client to %v: %v", serviceAddress, err)
	}
	c := obj.CreateClient(conn)
	return c, func() { conn.Close() }, nil
}

func PickNode(services []string) (string, error) {
	if len(services) <= 0 {
		return "", fmt.Errorf("no services available for service type")
	}

	index := rand.Intn(len(services))

	return services[index], nil
}

func (obj *ServiceClientBase[client_t]) ConnectMQ() (socket *zmq4.Socket, err error) {
	socket, err = zmq4.NewSocket(zmq4.REQ)
	if err != nil {
		return nil, err
	}
	endpoints, err := obj.getMQNodes()
	if err != nil {
		return nil, err
	}
	for _, endpoint := range endpoints {
		fmt.Printf("Connecting to %v\n", endpoint)
		err := socket.Connect(endpoint)
		if err != nil {
			return nil, err
		}
	}
	return socket, nil
}

func (obj *ServiceClientBase[client_t]) getMQNodes() ([]string, error) {
	registryClient := RegistryServiceClient.NewRegistryServiceClient(obj.RegistryAddresses)
	nodes, err := registryClient.Discover(obj.ServiceName + "MQ")
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
