package common

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/TAULargeScaleWorkshop/DTOY/config"
	RegistryServiceClient "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/client"
	. "github.com/TAULargeScaleWorkshop/DTOY/utils"
	"github.com/pebbe/zmq4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v2"
)

func LoadRegistryAddresses() []string {
	configFile := os.Args[1]
	configData, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error reading registry yaml file: %v", err)
		os.Exit(2)
	}
	var config config.ConfigBase
	err = yaml.Unmarshal(configData, &config) // parses YAML
	if err != nil {
		log.Fatalf("error unmarshaling registry addresses data: %v", err)
		os.Exit(3)
	}
	if len(config.RegistryAddresses) <= 0 {
		log.Fatalf("registry addresses yaml file does not include any enteries")
		os.Exit(4)
	}
	return config.RegistryAddresses
}

func startgRPC(listenPort int) (listeningAddress string, grpcServer *grpc.Server, startListening func(), portNum int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", listenPort))
	if err != nil {
		Logger.Fatalf("failed to listen: %v", err)
	}
	listeningAddress = lis.Addr().String()
	// get the port from lis
	port := lis.Addr().(*net.TCPAddr).Port
	grpcServer = grpc.NewServer()
	startListening = func() {
		if err := grpcServer.Serve(lis); err != nil {
			Logger.Fatalf("failed to serve: %v", err)
		}
	}
	return listeningAddress, grpcServer, startListening, port
}

func Start(serviceName string, grpcListenPort int, bindgRPCToService func(grpc.ServiceRegistrar), messageHandler func(method string, parameters []byte) (response proto.Message, err error)) (startListening func(), portNum int, registerService func() (unregister func())) {
	listeningAddressNonAsync, grpcServer, startListening, port := startgRPC(grpcListenPort)
	var startMQ func() = nil
	var listeningAddressAsync string
	if messageHandler != nil {
		fmt.Printf("Binding MQ to service\n")
		startMQ, listeningAddressAsync = bindMQToService(0, messageHandler)
	}

	bindgRPCToService(grpcServer)
	var unregisterFuncAsync func() = nil
	if messageHandler != nil {
		unregisterFuncAsyncc := registerAddress(serviceName+"MQ", LoadRegistryAddresses(), listeningAddressAsync)
		unregisterFuncAsync = func() {
			fmt.Printf("Unregistering MQ\n")
			unregisterFuncAsyncc()
		}
	}

	go func() {
		if unregisterFuncAsync != nil {
			defer unregisterFuncAsync()
		}
		if startMQ != nil {
			startMQ()
		}
	}()

	return startListening, port, func() (unregister func()) {
		return registerAddress(serviceName, LoadRegistryAddresses(), listeningAddressNonAsync)
	}
}

func registerAddress(serviceName string, registryAddresses []string,
	listeningAddress string) (unregister func()) {
	registryClient :=
		RegistryServiceClient.NewRegistryServiceClient(registryAddresses)
	err := registryClient.Register(serviceName, listeningAddress)
	if err != nil {
		Logger.Fatalf("Failed to register to registry service: %v", err)
	}
	return func() {
		registryClient.Unregister(serviceName, listeningAddress)
	}
}

func bindMQToService(listenPort int, messageHandler func(method string, parameters []byte) (response proto.Message, err error)) (startMQ func(), listeningAddress string) {

	socket, err := zmq4.NewSocket(zmq4.REP)
	if err != nil {
		Logger.Fatalf("Failed to create a new zmq socket: %v", err)
	}
	if listenPort == 0 {
		listeningAddress = "tcp://127.0.0.1:*"
	} else {
		listeningAddress = fmt.Sprintf("tcp://127.0.0.1:%v", listenPort)
	}
	err = socket.Bind(listeningAddress)
	if err != nil {
		Logger.Fatalf("Failed to bind a zmq socket: %v", err)
	}
	listeningAddress, err = socket.GetLastEndpoint()
	if err != nil {
		Logger.Fatalf("Failed to get listetning address of zmq socket: %v", err)
	}

	startMQ = func() {
		for {
			fmt.Printf("Waiting for data\n")
			data, readerr := socket.RecvBytes(0)
			if readerr != nil {
				Logger.Printf("Failed to receive bytes from MQ socket: %v\n", readerr)
				continue
			}
			if len(data) == 0 {
				continue
			}
			Logger.Printf("data len: %v\n", len(data))

			go func() {
				p := &CallParameters{}
				err := proto.Unmarshal(data, p)
				Logger.Printf("Marshal data: %v\n", p)
				payload := &ReturnValue{}

				if err != nil {
					Logger.Printf("Failed to unmarshal data: %v\n", err)
					payload.Error = err.Error()
					payload.Data = nil
					res, err := proto.Marshal(payload)
					if err != nil {
						Logger.Printf("Failed to marshal error response: %v\n", err)
						return
					}
					_, err = socket.SendBytes(res, 0)
					if err != nil {
						Logger.Printf("Failed to send error response: %v\n", err)
					}
					return
				}

				msgret, msgerr := messageHandler(p.Method, p.Data)
				msgMarshled, err := proto.Marshal(msgret)
				if err != nil {
					Logger.Printf("Failed to marshal response: %v\n", err)
					return
				}
				Logger.Printf("Marshal return value from function: %v\n", msgMarshled)
				if msgerr != nil {
					payload.Error = msgerr.Error()
				} else {
					payload.Error = ""
				}
				payload.Data = msgMarshled
				Logger.Printf("PreMarshal Payload %v\n", payload)
				res, err := proto.Marshal(payload)
				if err != nil {
					Logger.Printf("Failed to marshal response: %v\n", err)
					return
				}
				_, err = socket.SendBytes(res, 0)
				if err != nil {
					Logger.Printf("Failed to send response: %v\n", err)
				}
			}()
		}
	}

	return startMQ, listeningAddress
}

/*

	for {
		data, readerr := socket.RecvBytes(0)
		if err != nil {
			Logger.Printf("Failed to receive bytes from MQ socket: %v\n", readerr)
			continue
		}
		if len(data) == 0 {
			continue
		}
		Logger.Printf("data len: %v\n", len(data))

		go func() {
		}()
	}

*/
