package RegistryService

import (
	"context"
	"fmt"
	"net"

	"github.com/TAULargeScaleWorkshop/DTOY/config"
	. "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/common"
	RegistryServiceServant "github.com/TAULargeScaleWorkshop/DTOY/services/registry-service/servant"
	. "github.com/TAULargeScaleWorkshop/DTOY/utils"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v2"
)

type registryServiceImplementation struct {
	UnimplementedRegistryServiceServer
}

func startgRPC(listenPort int) (listeningAddress string, grpcServer *grpc.Server, startListening func(), err error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", listenPort))
	if err != nil {
		Logger.Printf("failed to listen: %v", err)
		return "", nil, nil, err
	}
	listeningAddress = lis.Addr().String()
	grpcServer = grpc.NewServer()
	startListening = func() {
		if err := grpcServer.Serve(lis); err != nil {
			Logger.Fatalf("failed to serve: %v", err)
		}
	}
	return listeningAddress, grpcServer, startListening, nil
}

func StartServer(baseGrpcListenPort int, grpcListenPort int, bindgRPCToService func(grpc.ServiceRegistrar)) error {

	listeningAddress, grpcServer, startListening, err := startgRPC(grpcListenPort)
	if err != nil {
		return err
	}
	_ = listeningAddress
	bindgRPCToService(grpcServer)
	Logger.Printf("RegistryService started at port %v\n", grpcListenPort)
	RegistryServiceServant.CreateChord(baseGrpcListenPort, grpcListenPort)
	if grpcListenPort == baseGrpcListenPort {
		go RegistryServiceServant.CheckIsAliveEvery10Seconds()
	}
	startListening()
	return nil
}

func Start(configData []byte) error {

	bindgRPCToService := func(s grpc.ServiceRegistrar) { RegisterRegistryServiceServer(s, &registryServiceImplementation{}) }

	var config config.RegistryConfigBase
	err := yaml.Unmarshal(configData, &config)

	if err != nil {
		Logger.Fatalf("error unmarshaling data: %v", err)
		return err
	}

	baseListenPort := config.ListenPort
	i := 0
	for {
		listenPort := baseListenPort + i
		err := StartServer(baseListenPort, listenPort, bindgRPCToService)
		if err == nil {
			break
		} else {
			i++
		}
	}
	return nil
}

func (obj *registryServiceImplementation) Register(ctx context.Context, in *ServiceRequest) (*empty.Empty, error) {
	err := RegistryServiceServant.Register(in.Name, in.Address)
	if err != nil {
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (obj *registryServiceImplementation) Unregister(ctx context.Context, in *ServiceRequest) (*empty.Empty, error) {
	err := RegistryServiceServant.Unregister(in.Name, in.Address)
	if err != nil {
		return &emptypb.Empty{}, nil
	}
	return &emptypb.Empty{}, nil
}

func (obj *registryServiceImplementation) Discover(ctx context.Context, in *wrapperspb.StringValue) (*ServiceNodes, error) {
	addresses, err := RegistryServiceServant.Discover(in.Value)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return &ServiceNodes{Nodes: []string{}}, nil
	}
	return &ServiceNodes{Nodes: addresses}, nil
}
