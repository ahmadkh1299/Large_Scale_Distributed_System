package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pebbe/zmq4"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ServiceClientBaseDirect struct {
	Address string
}

func NewServiceClientBaseDirect(address string) *ServiceClientBaseDirect {
	return &ServiceClientBaseDirect{Address: address}
}

func IsMessageQueueService(serviceName string) (bool){
	return strings.HasSuffix(serviceName, "MQ")
}

func (obj *ServiceClientBaseDirect) IsAlive(serviceName string) (res bool, err error) {
	if IsMessageQueueService(serviceName){
		socket, err := zmq4.NewSocket(zmq4.REP)
		if err != nil{
			return false, err
		}
		err = socket.Connect(obj.Address)
		if err != nil{
			return false, err
		}
		defer socket.Disconnect(obj.Address)
		return true, nil
	}else{
		fullMethodName := fmt.Sprintf("/%s.%s/IsAlive", strings.ToLower(serviceName), serviceName)
		conn, err := grpc.Dial(obj.Address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*5))
		if err != nil {
			return false, fmt.Errorf("failed to connect client to %v: %v", obj.Address, err)
		}
		out := new(wrappers.BoolValue)
		err = conn.Invoke(context.Background(), fullMethodName, &emptypb.Empty{}, out)
		if err != nil {
			return false, err
		}
		return out.Value, nil
	}
}

func (obj *ServiceClientBaseDirect) IsRoot() (res bool, err error) {
	conn, err := grpc.Dial(obj.Address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second*10))
	if err != nil {
		return false, fmt.Errorf("failed to connect client to %v: %v", obj.Address, err)
	}
	out := new(wrappers.BoolValue)
	err = conn.Invoke(context.Background(), "/cacheservice.CacheService/IsRoot", &emptypb.Empty{}, out)
	if err != nil {
		return false, err
	}
	return out.Value, nil
}
