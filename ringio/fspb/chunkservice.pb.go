// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: ringio/fspb/chunkservice.proto

package fspb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Key struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *Key) Reset() {
	*x = Key{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ringio_fspb_chunkservice_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Key) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Key) ProtoMessage() {}

func (x *Key) ProtoReflect() protoreflect.Message {
	mi := &file_ringio_fspb_chunkservice_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Key.ProtoReflect.Descriptor instead.
func (*Key) Descriptor() ([]byte, []int) {
	return file_ringio_fspb_chunkservice_proto_rawDescGZIP(), []int{0}
}

func (x *Key) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type PutRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key       *Key   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	ChunkName string `protobuf:"bytes,2,opt,name=chunk_name,json=chunkName,proto3" json:"chunk_name,omitempty"`
	ChunkSize int64  `protobuf:"varint,3,opt,name=chunk_size,json=chunkSize,proto3" json:"chunk_size,omitempty"`
	Data      []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *PutRequest) Reset() {
	*x = PutRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ringio_fspb_chunkservice_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRequest) ProtoMessage() {}

func (x *PutRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ringio_fspb_chunkservice_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRequest.ProtoReflect.Descriptor instead.
func (*PutRequest) Descriptor() ([]byte, []int) {
	return file_ringio_fspb_chunkservice_proto_rawDescGZIP(), []int{1}
}

func (x *PutRequest) GetKey() *Key {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *PutRequest) GetChunkName() string {
	if x != nil {
		return x.ChunkName
	}
	return ""
}

func (x *PutRequest) GetChunkSize() int64 {
	if x != nil {
		return x.ChunkSize
	}
	return 0
}

func (x *PutRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error     *Error         `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	Data      []byte         `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	ChunkInfo *HashChunkInfo `protobuf:"bytes,3,opt,name=chunk_info,json=chunkInfo,proto3" json:"chunk_info,omitempty"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ringio_fspb_chunkservice_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ringio_fspb_chunkservice_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_ringio_fspb_chunkservice_proto_rawDescGZIP(), []int{2}
}

func (x *GetResponse) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *GetResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *GetResponse) GetChunkInfo() *HashChunkInfo {
	if x != nil {
		return x.ChunkInfo
	}
	return nil
}

var File_ringio_fspb_chunkservice_proto protoreflect.FileDescriptor

var file_ringio_fspb_chunkservice_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x72, 0x69, 0x6e, 0x67, 0x69, 0x6f, 0x2f, 0x66, 0x73, 0x70, 0x62, 0x2f, 0x63, 0x68,
	0x75, 0x6e, 0x6b, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x04, 0x66, 0x73, 0x70, 0x62, 0x1a, 0x1b, 0x72, 0x69, 0x6e, 0x67, 0x69, 0x6f, 0x2f, 0x66,
	0x73, 0x70, 0x62, 0x2f, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x72, 0x69, 0x6e, 0x67, 0x69, 0x6f, 0x2f, 0x66, 0x73, 0x70, 0x62,
	0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x24, 0x72, 0x69,
	0x6e, 0x67, 0x69, 0x6f, 0x2f, 0x66, 0x73, 0x70, 0x62, 0x2f, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x2d,
	0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2d, 0x69, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x17, 0x0a, 0x03, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x7b, 0x0a, 0x0a, 0x50,
	0x75, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x4b, 0x65,
	0x79, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x68, 0x75, 0x6e,
	0x6b, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x73,
	0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x63, 0x68, 0x75, 0x6e, 0x6b,
	0x53, 0x69, 0x7a, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x78, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x32,
	0x0a, 0x0a, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x13, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x48, 0x61, 0x73, 0x68, 0x43, 0x68,
	0x75, 0x6e, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x09, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x49, 0x6e,
	0x66, 0x6f, 0x32, 0x9f, 0x03, 0x0a, 0x16, 0x48, 0x61, 0x73, 0x68, 0x43, 0x68, 0x75, 0x6e, 0x6b,
	0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x27, 0x0a,
	0x03, 0x47, 0x65, 0x74, 0x12, 0x09, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x4b, 0x65, 0x79, 0x1a,
	0x11, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x28, 0x0a, 0x03, 0x50, 0x75, 0x74, 0x12, 0x10, 0x2e,
	0x66, 0x73, 0x70, 0x62, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x0b, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x00, 0x28, 0x01,
	0x12, 0x22, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x09, 0x2e, 0x66, 0x73, 0x70,
	0x62, 0x2e, 0x4b, 0x65, 0x79, 0x1a, 0x0b, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x22, 0x00, 0x12, 0x36, 0x0a, 0x0a, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x12, 0x17, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0b, 0x2e, 0x66, 0x73,
	0x70, 0x62, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x00, 0x28, 0x01, 0x12, 0x35, 0x0a, 0x0a,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x12, 0x09, 0x2e, 0x66, 0x73, 0x70,
	0x62, 0x2e, 0x4b, 0x65, 0x79, 0x1a, 0x18, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x30, 0x01, 0x12, 0x29, 0x0a, 0x0d, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x12, 0x09, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x4b, 0x65, 0x79, 0x1a,
	0x0b, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x00, 0x12, 0x38,
	0x0a, 0x0c, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x12, 0x19,
	0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x69,
	0x63, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0b, 0x2e, 0x66, 0x73, 0x70, 0x62,
	0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x11, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x2e,
	0x66, 0x73, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x43, 0x68, 0x75, 0x6e,
	0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x0b, 0x2e, 0x66, 0x73, 0x70, 0x62, 0x2e, 0x45, 0x72, 0x72,
	0x6f, 0x72, 0x22, 0x00, 0x42, 0x18, 0x5a, 0x16, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x62, 0x6f, 0x72,
	0x61, 0x64, 0x2f, 0x72, 0x69, 0x6e, 0x67, 0x69, 0x6f, 0x2f, 0x66, 0x73, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ringio_fspb_chunkservice_proto_rawDescOnce sync.Once
	file_ringio_fspb_chunkservice_proto_rawDescData = file_ringio_fspb_chunkservice_proto_rawDesc
)

func file_ringio_fspb_chunkservice_proto_rawDescGZIP() []byte {
	file_ringio_fspb_chunkservice_proto_rawDescOnce.Do(func() {
		file_ringio_fspb_chunkservice_proto_rawDescData = protoimpl.X.CompressGZIP(file_ringio_fspb_chunkservice_proto_rawDescData)
	})
	return file_ringio_fspb_chunkservice_proto_rawDescData
}

var file_ringio_fspb_chunkservice_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_ringio_fspb_chunkservice_proto_goTypes = []interface{}{
	(*Key)(nil),                 // 0: fspb.Key
	(*PutRequest)(nil),          // 1: fspb.PutRequest
	(*GetResponse)(nil),         // 2: fspb.GetResponse
	(*Error)(nil),               // 3: fspb.Error
	(*HashChunkInfo)(nil),       // 4: fspb.HashChunkInfo
	(*PutReplicaRequest)(nil),   // 5: fspb.PutReplicaRequest
	(*CheckReplicaRequest)(nil), // 6: fspb.CheckReplicaRequest
	(*ReplicaChunkInfo)(nil),    // 7: fspb.ReplicaChunkInfo
	(*GetReplicaResponse)(nil),  // 8: fspb.GetReplicaResponse
}
var file_ringio_fspb_chunkservice_proto_depIdxs = []int32{
	0,  // 0: fspb.PutRequest.key:type_name -> fspb.Key
	3,  // 1: fspb.GetResponse.error:type_name -> fspb.Error
	4,  // 2: fspb.GetResponse.chunk_info:type_name -> fspb.HashChunkInfo
	0,  // 3: fspb.HashChunkSystemService.Get:input_type -> fspb.Key
	1,  // 4: fspb.HashChunkSystemService.Put:input_type -> fspb.PutRequest
	0,  // 5: fspb.HashChunkSystemService.Delete:input_type -> fspb.Key
	5,  // 6: fspb.HashChunkSystemService.PutReplica:input_type -> fspb.PutReplicaRequest
	0,  // 7: fspb.HashChunkSystemService.GetReplica:input_type -> fspb.Key
	0,  // 8: fspb.HashChunkSystemService.DeleteReplica:input_type -> fspb.Key
	6,  // 9: fspb.HashChunkSystemService.CheckReplica:input_type -> fspb.CheckReplicaRequest
	7,  // 10: fspb.HashChunkSystemService.UpdateReplicaInfo:input_type -> fspb.ReplicaChunkInfo
	2,  // 11: fspb.HashChunkSystemService.Get:output_type -> fspb.GetResponse
	3,  // 12: fspb.HashChunkSystemService.Put:output_type -> fspb.Error
	3,  // 13: fspb.HashChunkSystemService.Delete:output_type -> fspb.Error
	3,  // 14: fspb.HashChunkSystemService.PutReplica:output_type -> fspb.Error
	8,  // 15: fspb.HashChunkSystemService.GetReplica:output_type -> fspb.GetReplicaResponse
	3,  // 16: fspb.HashChunkSystemService.DeleteReplica:output_type -> fspb.Error
	3,  // 17: fspb.HashChunkSystemService.CheckReplica:output_type -> fspb.Error
	3,  // 18: fspb.HashChunkSystemService.UpdateReplicaInfo:output_type -> fspb.Error
	11, // [11:19] is the sub-list for method output_type
	3,  // [3:11] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_ringio_fspb_chunkservice_proto_init() }
func file_ringio_fspb_chunkservice_proto_init() {
	if File_ringio_fspb_chunkservice_proto != nil {
		return
	}
	file_ringio_fspb_chunkinfo_proto_init()
	file_ringio_fspb_error_proto_init()
	file_ringio_fspb_chunk_replica_info_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ringio_fspb_chunkservice_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Key); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ringio_fspb_chunkservice_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ringio_fspb_chunkservice_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ringio_fspb_chunkservice_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ringio_fspb_chunkservice_proto_goTypes,
		DependencyIndexes: file_ringio_fspb_chunkservice_proto_depIdxs,
		MessageInfos:      file_ringio_fspb_chunkservice_proto_msgTypes,
	}.Build()
	File_ringio_fspb_chunkservice_proto = out.File
	file_ringio_fspb_chunkservice_proto_rawDesc = nil
	file_ringio_fspb_chunkservice_proto_goTypes = nil
	file_ringio_fspb_chunkservice_proto_depIdxs = nil
}
