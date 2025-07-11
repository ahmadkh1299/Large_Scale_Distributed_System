package CacheServiceClient

import (
	context "context"
	"fmt"

	service "github.com/TAULargeScaleWorkshop/DTOY/services/cache-service/common"
	services "github.com/TAULargeScaleWorkshop/DTOY/services/common"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type CacheServiceClient struct {
	services.ServiceClientBase[service.CacheServiceClient]
}

func NewCacheServiceClient(RegistryAddresses []string) *CacheServiceClient {
	client := &CacheServiceClient{

		ServiceClientBase: services.ServiceClientBase[service.CacheServiceClient]{
			RegistryAddresses: RegistryAddresses,
			ServiceName:       "CacheService",
			CreateClient:      service.NewCacheServiceClient,
		},
	}
	if RegistryAddresses == nil {
		client.ServiceClientBase.LoadRegistryAddresses()
	}
	return client
}

func (obj *CacheServiceClient) IsAlive() (bool, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return false, fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()

	r, err := c.IsAlive(context.Background(), &emptypb.Empty{})
	if err != nil {
		return false, fmt.Errorf("could not call IsAlive: %v", err)
	}
	return r.Value, nil
}

func (obj *CacheServiceClient) Set(key string, value string) error {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()

	_, err = c.Set(context.Background(), &service.SetRequest{Key: key, Value: value})
	if err != nil {
		return fmt.Errorf("could not call IsAlive: %v", err)
	}
	return nil
}

func (obj *CacheServiceClient) Get(key string) (string, error) {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()

	res, err := c.Get(context.Background(), &wrapperspb.StringValue{Value: key})
	if err != nil {
		return "", fmt.Errorf("could not call IsAlive: %v", err)
	}
	return res.Value, nil
}

func (obj *CacheServiceClient) Delete(key string) error {
	c, closeFunc, err := obj.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect. Error: %v", err)
	}
	defer closeFunc()

	_, err = c.Delete(context.Background(), &wrapperspb.StringValue{Value: key})
	if err != nil {
		return fmt.Errorf("could not call IsAlive: %v", err)
	}
	return nil
}
