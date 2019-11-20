// Code generated by protoc-gen-go. DO NOT EDIT.
// source: kv.proto

package baetyl

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// KV kv message
type KV struct {
	// key is the key, in bytes, to put into the key-value store.
	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// value is the value, in bytes, to associate with the key in the key-value store.
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KV) Reset()         { *m = KV{} }
func (m *KV) String() string { return proto.CompactTextString(m) }
func (*KV) ProtoMessage()    {}
func (*KV) Descriptor() ([]byte, []int) {
	return fileDescriptor_2216fe83c9c12408, []int{0}
}

func (m *KV) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KV.Unmarshal(m, b)
}
func (m *KV) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KV.Marshal(b, m, deterministic)
}
func (m *KV) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KV.Merge(m, src)
}
func (m *KV) XXX_Size() int {
	return xxx_messageInfo_KV.Size(m)
}
func (m *KV) XXX_DiscardUnknown() {
	xxx_messageInfo_KV.DiscardUnknown(m)
}

var xxx_messageInfo_KV proto.InternalMessageInfo

func (m *KV) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *KV) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type KVs struct {
	Kvs                  []*KV    `protobuf:"bytes,1,rep,name=kvs,proto3" json:"kvs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *KVs) Reset()         { *m = KVs{} }
func (m *KVs) String() string { return proto.CompactTextString(m) }
func (*KVs) ProtoMessage()    {}
func (*KVs) Descriptor() ([]byte, []int) {
	return fileDescriptor_2216fe83c9c12408, []int{1}
}

func (m *KVs) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_KVs.Unmarshal(m, b)
}
func (m *KVs) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_KVs.Marshal(b, m, deterministic)
}
func (m *KVs) XXX_Merge(src proto.Message) {
	xxx_messageInfo_KVs.Merge(m, src)
}
func (m *KVs) XXX_Size() int {
	return xxx_messageInfo_KVs.Size(m)
}
func (m *KVs) XXX_DiscardUnknown() {
	xxx_messageInfo_KVs.DiscardUnknown(m)
}

var xxx_messageInfo_KVs proto.InternalMessageInfo

func (m *KVs) GetKvs() []*KV {
	if m != nil {
		return m.Kvs
	}
	return nil
}

func init() {
	proto.RegisterType((*KV)(nil), "baetyl.KV")
	proto.RegisterType((*KVs)(nil), "baetyl.KVs")
}

func init() { proto.RegisterFile("kv.proto", fileDescriptor_2216fe83c9c12408) }

var fileDescriptor_2216fe83c9c12408 = []byte{
	// 168 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xc8, 0x2e, 0xd3, 0x2b,
	0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4b, 0x4a, 0x4c, 0x2d, 0xa9, 0xcc, 0x51, 0xd2, 0xe1, 0x62,
	0xf2, 0x0e, 0x13, 0x12, 0xe0, 0x62, 0xce, 0x4e, 0xad, 0x94, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x09,
	0x02, 0x31, 0x85, 0x44, 0xb8, 0x58, 0xcb, 0x12, 0x73, 0x4a, 0x53, 0x25, 0x98, 0xc0, 0x62, 0x10,
	0x8e, 0x92, 0x32, 0x17, 0xb3, 0x77, 0x58, 0xb1, 0x90, 0x0c, 0x17, 0x73, 0x76, 0x59, 0xb1, 0x04,
	0xa3, 0x02, 0xb3, 0x06, 0xb7, 0x11, 0x97, 0x1e, 0xc4, 0x28, 0x3d, 0xef, 0xb0, 0x20, 0x90, 0xb0,
	0xd1, 0x44, 0x46, 0x2e, 0x4e, 0xef, 0xb0, 0xe0, 0xd4, 0xa2, 0xb2, 0xcc, 0xe4, 0x54, 0x21, 0x79,
	0x2e, 0xe6, 0xe0, 0xd4, 0x12, 0x21, 0x24, 0x55, 0x52, 0x48, 0x6c, 0x25, 0x06, 0x90, 0x02, 0x77,
	0x42, 0x0a, 0x5c, 0x52, 0x73, 0xf0, 0x28, 0x50, 0xe4, 0x62, 0xf1, 0xc9, 0x2c, 0x46, 0x35, 0x82,
	0x1b, 0xc1, 0x2e, 0x56, 0x62, 0x48, 0x62, 0x03, 0xfb, 0xda, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff,
	0xff, 0xae, 0x05, 0x11, 0x01, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// KVServiceClient is the client API for KVService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KVServiceClient interface {
	Set(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KV, error)
	Get(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KV, error)
	Del(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KV, error)
	List(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KVs, error)
}

type kVServiceClient struct {
	cc *grpc.ClientConn
}

func NewKVServiceClient(cc *grpc.ClientConn) KVServiceClient {
	return &kVServiceClient{cc}
}

func (c *kVServiceClient) Set(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KV, error) {
	out := new(KV)
	err := c.cc.Invoke(ctx, "/baetyl.KVService/Set", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) Get(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KV, error) {
	out := new(KV)
	err := c.cc.Invoke(ctx, "/baetyl.KVService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) Del(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KV, error) {
	out := new(KV)
	err := c.cc.Invoke(ctx, "/baetyl.KVService/Del", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *kVServiceClient) List(ctx context.Context, in *KV, opts ...grpc.CallOption) (*KVs, error) {
	out := new(KVs)
	err := c.cc.Invoke(ctx, "/baetyl.KVService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KVServiceServer is the server API for KVService service.
type KVServiceServer interface {
	Set(context.Context, *KV) (*KV, error)
	Get(context.Context, *KV) (*KV, error)
	Del(context.Context, *KV) (*KV, error)
	List(context.Context, *KV) (*KVs, error)
}

// UnimplementedKVServiceServer can be embedded to have forward compatible implementations.
type UnimplementedKVServiceServer struct {
}

func (*UnimplementedKVServiceServer) Set(ctx context.Context, req *KV) (*KV, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Set not implemented")
}
func (*UnimplementedKVServiceServer) Get(ctx context.Context, req *KV) (*KV, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedKVServiceServer) Del(ctx context.Context, req *KV) (*KV, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Del not implemented")
}
func (*UnimplementedKVServiceServer) List(ctx context.Context, req *KV) (*KVs, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}

func RegisterKVServiceServer(s *grpc.Server, srv KVServiceServer) {
	s.RegisterService(&_KVService_serviceDesc, srv)
}

func _KVService_Set_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KV)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).Set(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/baetyl.KVService/Set",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).Set(ctx, req.(*KV))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KV)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/baetyl.KVService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).Get(ctx, req.(*KV))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_Del_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KV)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).Del(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/baetyl.KVService/Del",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).Del(ctx, req.(*KV))
	}
	return interceptor(ctx, in, info, handler)
}

func _KVService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KV)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KVServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/baetyl.KVService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KVServiceServer).List(ctx, req.(*KV))
	}
	return interceptor(ctx, in, info, handler)
}

var _KVService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "baetyl.KVService",
	HandlerType: (*KVServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Set",
			Handler:    _KVService_Set_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _KVService_Get_Handler,
		},
		{
			MethodName: "Del",
			Handler:    _KVService_Del_Handler,
		},
		{
			MethodName: "List",
			Handler:    _KVService_List_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kv.proto",
}
