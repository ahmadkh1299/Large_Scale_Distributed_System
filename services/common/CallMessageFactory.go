package common

import (
	"google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
)


func NewMarshaledCallParameter(method string, kwargs ...protoreflect.ProtoMessage) ([]byte, error){
	payload := &CallParameters{}
	payload.Method = method

	byteArr := make([]byte, 0)
	for _, arg := range kwargs{
		marshledArg, err := proto.Marshal(arg)
		if err != nil {
			return nil, err
		}
		byteArr = append(byteArr, marshledArg...)
	}
	payload.Data = byteArr
	marshledPayload, err := proto.Marshal(payload)
	if err != nil{
		return nil, err
	}
	return marshledPayload, nil
}

func UnmarshalReturnValue(rv []byte) (*ReturnValue, error) {
	ret := &ReturnValue{}
	err := proto.Unmarshal(rv, ret)
	if err != nil{
		return nil, err
	}
	return ret, nil
}

func (rv *ReturnValue) ExtractInnerMessage(p protoreflect.ProtoMessage) (error) {
	err := proto.Unmarshal(rv.Data, p)
	if err != nil{
		return err
	}
	return nil
}

func ParseParamsIntoBytes(kwargs ...protoreflect.ProtoMessage) ([]byte, error){
	byteArr := make([]byte, 0)
	for _, arg := range kwargs{
		marshledArg, err := proto.Marshal(arg)
		if err != nil {
			return nil, err
		}
		byteArr = append(byteArr, marshledArg...)
	}
	return byteArr, nil
}