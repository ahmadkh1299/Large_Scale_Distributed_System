package TestServiceClient

import (
	context "context"
	"errors"
	"fmt"

	services "github.com/TAULargeScaleWorkshop/DTOY/services/common"
	service "github.com/TAULargeScaleWorkshop/DTOY/services/test-service/common"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type TestServiceClient struct {
	services.ServiceClientBase[service.TestServiceClient]
}

func NewTestServiceClient() *TestServiceClient {
	client := &TestServiceClient{

		ServiceClientBase: services.ServiceClientBase[service.TestServiceClient]{
			ServiceName:  "TestService",
			CreateClient: service.NewTestServiceClient,
		},
	}
	client.ServiceClientBase.LoadRegistryAddresses()
	return client
}

func (obj *TestServiceClient) HelloWorld() (string, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()
	// Call the HelloWorld RPC function
	r, err := c.HelloWorld(context.Background(), &emptypb.Empty{})
	if err != nil {
		return "", fmt.Errorf("could not call HelloWorld: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) HelloToUser(name string) (string, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()
	// Call the HelloWorld RPC function
	r, err := c.HelloToUser(context.Background(), wrapperspb.String(name))
	if err != nil {
		return "", fmt.Errorf("could not call HelloToUser: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) Store(key string, value string) error {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()
	// Call the HelloWorld RPC function
	_, storeErr := c.Store(context.Background(), &service.StoreKeyValue{Key: key, Value: value})
	if storeErr != nil {
		return fmt.Errorf("could not call HelloToUser: %v", err)
	}
	return nil
}

func (obj *TestServiceClient) Get(key string) (string, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()
	// Call the HelloWorld RPC function
	r, err := c.Get(context.Background(), wrapperspb.String(key))
	if err != nil {
		return "", fmt.Errorf("could not call HelloToUser: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) WaitAndRand(seconds int32) (func() (int32, error), error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect. Error: %v", err)
	}
	r, err := c.WaitAndRand(context.Background(), wrapperspb.Int32(seconds))
	if err != nil {
		return nil, fmt.Errorf("could not call Get: %v", err)
	}
	res := func() (int32, error) {
		defer closeFunc()
		x, err := r.Recv()
		return x.Value, err
	}
	return res, nil
}

func (obj *TestServiceClient) IsAlive() (bool, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return false, fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()

	r, err := c.IsAlive(context.Background(), &emptypb.Empty{})
	if err != nil {
		return false, fmt.Errorf("could not call HelloToUser: %v", err)
	}
	return r.Value, nil
}

func (obj *TestServiceClient) ExtractLinksFromURL(url string, depth int32) ([]string, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()

	r, err := c.ExtractLinksFromURL(context.Background(), &service.ExtractLinksFromURLParameters{Url: url, Depth: depth})
	if err != nil {
		return nil, fmt.Errorf("could not call HelloToUser: %v", err)
	}
	return r.Links, nil
}

func (obj *TestServiceClient) HelloWorldAsync() (func() (string, error), error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, err
	}
	msg, err := services.NewMarshaledCallParameter("HelloWorld", &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		return nil, err
	}

	// return function (future pattern)
	ret := func() (string, error) {
		defer mqsocket.Close()
		rvBytes, err := mqsocket.RecvBytes(0)
		if err != nil {
			return "", err
		}
		rv, err := services.UnmarshalReturnValue(rvBytes)
		if err != nil {
			return "", err
		}
		if rv.Error != "" {
			return "", errors.New(rv.Error)
		}
		str := &wrapperspb.StringValue{}
		err = rv.ExtractInnerMessage(str)
		if err != nil {
			return "", err
		}
		return str.Value, nil
	}

	return ret, nil
}

func (obj *TestServiceClient) HelloToUserAsync(user string) (func() (string, error), error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, err
	}
	msg, err := services.NewMarshaledCallParameter("HelloToUser", &wrappers.StringValue{Value: user})
	if err != nil {
		return nil, err
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		return nil, err
	}

	// return function (future pattern)
	ret := func() (string, error) {
		defer mqsocket.Close()
		rvBytes, err := mqsocket.RecvBytes(0)
		if err != nil {
			return "", err
		}
		rv, err := services.UnmarshalReturnValue(rvBytes)
		if err != nil {
			return "", err
		}
		if rv.Error != "" {
			return "", errors.New(rv.Error)
		}
		str := &wrapperspb.StringValue{}
		err = rv.ExtractInnerMessage(str)
		if err != nil {
			return "", err
		}
		return str.Value, nil
	}

	return ret, nil
}

func (obj *TestServiceClient) GetAsync(key string) (func() (string, error), error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, err
	}
	msg, err := services.NewMarshaledCallParameter("Get", &wrappers.StringValue{Value: key})
	if err != nil {
		return nil, err
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		return nil, err
	}

	// return function (future pattern)
	ret := func() (string, error) {
		defer mqsocket.Close()
		rvBytes, err := mqsocket.RecvBytes(0)
		if err != nil {
			return "", err
		}
		rv, err := services.UnmarshalReturnValue(rvBytes)
		if err != nil {
			return "", err
		}
		if rv.Error != "" {
			return "", errors.New(rv.Error)
		}
		str := &wrappers.StringValue{}
		err = rv.ExtractInnerMessage(str)
		if err != nil {
			return "", err
		}
		return str.Value, nil
	}

	return ret, nil
}

func (obj *TestServiceClient) ExtractLinksFromURLAsync(url string, depth int32) (func() ([]string, error), error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, err
	}
	msg, err := services.NewMarshaledCallParameter("ExtractLinksFromURL", &service.ExtractLinksFromURLParameters{Url: url, Depth: depth})
	if err != nil {
		fmt.Printf("Im here 1\n")
		return nil, err
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		fmt.Printf("Im here 2\n")
		return nil, err
	}

	// return function (future pattern)
	ret := func() ([]string, error) {
		defer mqsocket.Close()
		rvBytes, err := mqsocket.RecvBytes(0)
		if err != nil {
			fmt.Printf("Im here 3\n")
			return nil, err
		}
		rv, err := services.UnmarshalReturnValue(rvBytes)
		if err != nil {
			fmt.Printf("Im here 4\n")
			return nil, err
		}
		if rv.Error != "" {
			fmt.Printf("Im here 5\n")
			return nil, errors.New(rv.Error)
		}
		ret := &service.ExtractLinksFromURLReturnedValue{}
		err = rv.ExtractInnerMessage(ret)
		if err != nil {
			fmt.Printf("Im here 6\n")
			return nil, err
		}
		return ret.Links, nil
	}

	return ret, nil
}

func (obj *TestServiceClient) StoreAsync(key string, value string) (func() error, error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, err
	}
	msg, err := services.NewMarshaledCallParameter("Store", &service.StoreKeyValue{Key: key, Value: value})
	if err != nil {
		return nil, err
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		return nil, err
	}

	// return function (future pattern)
	ret := func() error {
		defer mqsocket.Close()
		rvBytes, err := mqsocket.RecvBytes(0)
		if err != nil {
			return err
		}
		rv, err := services.UnmarshalReturnValue(rvBytes)
		if err != nil {
			return err
		}
		if rv.Error != "" {
			return errors.New(rv.Error)
		}
		str := &empty.Empty{}
		err = rv.ExtractInnerMessage(str)
		if err != nil {
			return err
		}
		return nil
	}

	return ret, nil
}

func (obj *TestServiceClient) IsAliveAsync(key string) (func() (bool, error), error) {
	mqsocket, err := obj.ConnectMQ()
	if err != nil {
		return nil, err
	}
	msg, err := services.NewMarshaledCallParameter("IsAlive", &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	_, err = mqsocket.SendBytes(msg, 0)
	if err != nil {
		return nil, err
	}

	// return function (future pattern)
	ret := func() (bool, error) {
		defer mqsocket.Close()
		rvBytes, err := mqsocket.RecvBytes(0)
		if err != nil {
			return false, err
		}
		rv, err := services.UnmarshalReturnValue(rvBytes)
		if err != nil {
			return false, err
		}
		if rv.Error != "" {
			return false, errors.New(rv.Error)
		}
		boolean := &wrappers.BoolValue{}
		err = rv.ExtractInnerMessage(boolean)
		if err != nil {
			return false, err
		}
		return boolean.Value, nil
	}

	return ret, nil
}
