package CacheService

import (
	"context"

	. "github.com/TAULargeScaleWorkshop/DTOY/services/cache-service/common"
	CacheServiceServant "github.com/TAULargeScaleWorkshop/DTOY/services/cache-service/servant"
	services "github.com/TAULargeScaleWorkshop/DTOY/services/common"
	"github.com/TAULargeScaleWorkshop/DTOY/utils"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type cacheServiceImplementation struct {
	UnimplementedCacheServiceServer
}

func Start(configData []byte) error {
	bindgRPCToService := func(s grpc.ServiceRegistrar) { RegisterCacheServiceServer(s, &cacheServiceImplementation{}) }
	startListening, port, register := services.Start("CacheService", 0, bindgRPCToService, nil)
	CacheServiceServant.CreateChord(port)
	unregister := register()
	defer unregister()
	utils.Logger.Printf("CacheService server started on port: %v\n", port)
	startListening()
	return nil
}

func (obj *cacheServiceImplementation) Get(ctxt context.Context, key *wrappers.StringValue) (*wrappers.StringValue, error) {
	val, err := CacheServiceServant.Get(key.Value)
	if err != nil {
		return nil, err
	}
	return wrapperspb.String(val), nil
}

func (obj *cacheServiceImplementation) Set(ctxt context.Context, keyvalue *SetRequest) (*empty.Empty, error) {
	CacheServiceServant.Set(keyvalue.Key, keyvalue.Value)
	return &emptypb.Empty{}, nil
}

func (obj *cacheServiceImplementation) Delete(ctxt context.Context, key *wrappers.StringValue) (*empty.Empty, error) {
	CacheServiceServant.Delete(key.Value)
	return &emptypb.Empty{}, nil
}

func (obj *cacheServiceImplementation) IsAlive(ctxt context.Context, _ *empty.Empty) (*wrappers.BoolValue, error) {
	return wrapperspb.Bool(true), nil
}

func (obj *cacheServiceImplementation) IsRoot(context.Context, *empty.Empty) (*wrappers.BoolValue, error) {
	val, err := CacheServiceServant.IsRoot()
	if err != nil {
		return nil, err
	}
	return wrapperspb.Bool(val), nil
}
