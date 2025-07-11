package TestService

import (
	"context"
	"fmt"
	_ "log"
	_ "net"

	services "github.com/TAULargeScaleWorkshop/DTOY/services/common"
	. "github.com/TAULargeScaleWorkshop/DTOY/services/test-service/common"
	TestServiceServant "github.com/TAULargeScaleWorkshop/DTOY/services/test-service/servant"
	"github.com/TAULargeScaleWorkshop/DTOY/utils"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type testServiceImplementation struct {
	UnimplementedTestServiceServer
}

var serviceInstance *testServiceImplementation

func messageHandler(method string, parameters []byte) (response proto.Message, err error) {
	switch method {
	case "ExtractLinksFromURL":
		p := &ExtractLinksFromURLParameters{}
		err := proto.Unmarshal(parameters, p)

		if err != nil {
			return nil, err
		}

		res, err := serviceInstance.ExtractLinksFromURL(context.Background(), p)

		if err != nil {
			return nil, err
		}
		return res, nil
	case "HelloWorld":
		p := &empty.Empty{}
		err := proto.Unmarshal(parameters, p)

		if err != nil {
			return nil, err
		}

		res, err := serviceInstance.HelloWorld(context.Background(), p)

		if err != nil {
			return nil, err
		}
		return res, nil
	case "HelloToUser":
		p := &wrappers.StringValue{}
		err := proto.Unmarshal(parameters, p)

		if err != nil {
			return nil, err
		}

		res, err := serviceInstance.HelloToUser(context.Background(), p)

		if err != nil {
			return nil, err
		}
		return res, nil
	case "Store":
		p := &StoreKeyValue{}
		err := proto.Unmarshal(parameters, p)

		if err != nil {
			return nil, err
		}

		res, err := serviceInstance.Store(context.Background(), p)

		if err != nil {
			return nil, err
		}
		return res, nil
	case "Get":
		p := &wrappers.StringValue{}
		err := proto.Unmarshal(parameters, p)

		if err != nil {
			return nil, err
		}

		res, err := serviceInstance.Get(context.Background(), p)

		if err != nil {
			return nil, err
		}
		return res, nil
	case "IsAlive":
		p := &empty.Empty{}
		err := proto.Unmarshal(parameters, p)

		if err != nil {
			return nil, err
		}

		res, err := serviceInstance.IsAlive(context.Background(), p)

		if err != nil {
			return nil, err
		}
		return res, nil
	default:
		return nil, fmt.Errorf("MQ message called unknown method: %v", method)
	}
}

func Start(configData []byte) error {
	serviceInstance = &testServiceImplementation{}
	bindgRPCToService := func(s grpc.ServiceRegistrar) { RegisterTestServiceServer(s, serviceInstance) }
	startListening, port, register := services.Start("TestService", 0, bindgRPCToService, messageHandler)
	unregister := register()
	defer unregister()
	utils.Logger.Printf("TestService server started on port %v\n", port)
	startListening()
	return nil
}

func (obj *testServiceImplementation) HelloWorld(ctxt context.Context, _ *emptypb.Empty) (res *wrapperspb.StringValue, err error) {
	return wrapperspb.String(TestServiceServant.HelloWorld()), nil
}

func (obj *testServiceImplementation) HelloToUser(ctxt context.Context, name *wrappers.StringValue) (res *wrapperspb.StringValue, err error) {
	return wrapperspb.String(TestServiceServant.HelloToUser(name.Value)), nil
}

func (testServiceImplementation) Store(ctxt context.Context, keyvalue *StoreKeyValue) (*empty.Empty, error) {
	err := TestServiceServant.Store(keyvalue.Key, keyvalue.Value)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (testServiceImplementation) Get(ctxt context.Context, key *wrappers.StringValue) (*wrappers.StringValue, error) {
	value, err := TestServiceServant.Get(key.Value)
	if err != nil {
		return &wrapperspb.StringValue{}, err
	}
	return wrapperspb.String(value), nil
}

func (obj *testServiceImplementation) WaitAndRand(seconds *wrapperspb.Int32Value, streamRet TestService_WaitAndRandServer) error {
	utils.Logger.Printf("WaitAndRand called")
	streamClient := func(x int32) error {
		return streamRet.Send(wrapperspb.Int32(x))
	}
	return TestServiceServant.WaitAndRand(seconds.Value, streamClient)
}

func (testServiceImplementation) IsAlive(ctxt context.Context, _ *empty.Empty) (*wrappers.BoolValue, error) {
	return wrapperspb.Bool(true), nil
}

func (testServiceImplementation) ExtractLinksFromURL(ctxt context.Context, links *ExtractLinksFromURLParameters) (*ExtractLinksFromURLReturnedValue, error) {
	linkArr, err := TestServiceServant.ExtractLinksFromURL(links.Url, links.Depth)
	if err != nil {
		return &ExtractLinksFromURLReturnedValue{}, err
	}
	return &ExtractLinksFromURLReturnedValue{Links: linkArr}, nil
}
