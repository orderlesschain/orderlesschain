// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.12
// source: sequencer.proto

package protos

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

type SequencedTransactionForOrdering struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId string `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
}

func (x *SequencedTransactionForOrdering) Reset() {
	*x = SequencedTransactionForOrdering{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sequencer_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SequencedTransactionForOrdering) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SequencedTransactionForOrdering) ProtoMessage() {}

func (x *SequencedTransactionForOrdering) ProtoReflect() protoreflect.Message {
	mi := &file_sequencer_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SequencedTransactionForOrdering.ProtoReflect.Descriptor instead.
func (*SequencedTransactionForOrdering) Descriptor() ([]byte, []int) {
	return file_sequencer_proto_rawDescGZIP(), []int{0}
}

func (x *SequencedTransactionForOrdering) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

type SequencedTransactionForCommitting struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId string `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
}

func (x *SequencedTransactionForCommitting) Reset() {
	*x = SequencedTransactionForCommitting{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sequencer_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SequencedTransactionForCommitting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SequencedTransactionForCommitting) ProtoMessage() {}

func (x *SequencedTransactionForCommitting) ProtoReflect() protoreflect.Message {
	mi := &file_sequencer_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SequencedTransactionForCommitting.ProtoReflect.Descriptor instead.
func (*SequencedTransactionForCommitting) Descriptor() ([]byte, []int) {
	return file_sequencer_proto_rawDescGZIP(), []int{1}
}

func (x *SequencedTransactionForCommitting) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

var File_sequencer_proto protoreflect.FileDescriptor

var file_sequencer_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x1a, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0c, 0x73, 0x68,
	0x61, 0x72, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3a, 0x0a, 0x1f, 0x53, 0x65,
	0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x46, 0x6f, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x17, 0x0a,
	0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x22, 0x3c, 0x0a, 0x21, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e,
	0x63, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x6f,
	0x72, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x12, 0x17, 0x0a, 0x07, 0x6e,
	0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f,
	0x64, 0x65, 0x49, 0x64, 0x32, 0xf5, 0x03, 0x0a, 0x10, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63,
	0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3c, 0x0a, 0x10, 0x42, 0x49, 0x44,
	0x4c, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x17, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x28, 0x01, 0x12, 0x66, 0x0a, 0x20, 0x53, 0x75, 0x62, 0x73, 0x63,
	0x72, 0x69, 0x62, 0x65, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x46, 0x6f, 0x72, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x27, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x64, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x6f, 0x72, 0x4f, 0x72, 0x64, 0x65,
	0x72, 0x69, 0x6e, 0x67, 0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x50, 0x72,
	0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x30, 0x01, 0x12,
	0x6a, 0x0a, 0x22, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x54, 0x72, 0x61, 0x6e,
	0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x46, 0x6f, 0x72, 0x50, 0x72, 0x6f, 0x63, 0x65,
	0x73, 0x73, 0x69, 0x6e, 0x67, 0x12, 0x29, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x53,
	0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x46, 0x6f, 0x72, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x74, 0x69, 0x6e, 0x67,
	0x1a, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x61, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x30, 0x01, 0x12, 0x39, 0x0a, 0x11, 0x43,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x12, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x1a, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x4a, 0x0a, 0x1d, 0x47, 0x65, 0x74, 0x54, 0x72, 0x61,
	0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e,
	0x4c, 0x61, 0x74, 0x65, 0x6e, 0x63, 0x79, 0x42, 0x72, 0x65, 0x61, 0x6b, 0x44, 0x6f, 0x77, 0x6e,
	0x30, 0x01, 0x12, 0x48, 0x0a, 0x1b, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x4d, 0x65, 0x6d, 0x6f,
	0x72, 0x79, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x0d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x1a, 0x18, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x6f, 0x72, 0x79,
	0x43, 0x50, 0x55, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x30, 0x01, 0x42, 0x0b, 0x5a, 0x09,
	0x2e, 0x2f, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_sequencer_proto_rawDescOnce sync.Once
	file_sequencer_proto_rawDescData = file_sequencer_proto_rawDesc
)

func file_sequencer_proto_rawDescGZIP() []byte {
	file_sequencer_proto_rawDescOnce.Do(func() {
		file_sequencer_proto_rawDescData = protoimpl.X.CompressGZIP(file_sequencer_proto_rawDescData)
	})
	return file_sequencer_proto_rawDescData
}

var file_sequencer_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_sequencer_proto_goTypes = []interface{}{
	(*SequencedTransactionForOrdering)(nil),   // 0: protos.SequencedTransactionForOrdering
	(*SequencedTransactionForCommitting)(nil), // 1: protos.SequencedTransactionForCommitting
	(*ProposalRequest)(nil),                   // 2: protos.ProposalRequest
	(*OperationMode)(nil),                     // 3: protos.OperationMode
	(*Empty)(nil),                             // 4: protos.Empty
	(*LatencyBreakDown)(nil),                  // 5: protos.LatencyBreakDown
	(*MemoryCPUProfile)(nil),                  // 6: protos.MemoryCPUProfile
}
var file_sequencer_proto_depIdxs = []int32{
	2, // 0: protos.SequencerService.BIDLTransactions:input_type -> protos.ProposalRequest
	0, // 1: protos.SequencerService.SubscribeTransactionsForOrdering:input_type -> protos.SequencedTransactionForOrdering
	1, // 2: protos.SequencerService.SubscribeTransactionsForProcessing:input_type -> protos.SequencedTransactionForCommitting
	3, // 3: protos.SequencerService.ChangeModeRestart:input_type -> protos.OperationMode
	4, // 4: protos.SequencerService.GetTransactionProfilingResult:input_type -> protos.Empty
	4, // 5: protos.SequencerService.GetCPUMemoryProfilingResult:input_type -> protos.Empty
	4, // 6: protos.SequencerService.BIDLTransactions:output_type -> protos.Empty
	2, // 7: protos.SequencerService.SubscribeTransactionsForOrdering:output_type -> protos.ProposalRequest
	2, // 8: protos.SequencerService.SubscribeTransactionsForProcessing:output_type -> protos.ProposalRequest
	4, // 9: protos.SequencerService.ChangeModeRestart:output_type -> protos.Empty
	5, // 10: protos.SequencerService.GetTransactionProfilingResult:output_type -> protos.LatencyBreakDown
	6, // 11: protos.SequencerService.GetCPUMemoryProfilingResult:output_type -> protos.MemoryCPUProfile
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_sequencer_proto_init() }
func file_sequencer_proto_init() {
	if File_sequencer_proto != nil {
		return
	}
	file_transaction_proto_init()
	file_shared_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_sequencer_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SequencedTransactionForOrdering); i {
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
		file_sequencer_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SequencedTransactionForCommitting); i {
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
			RawDescriptor: file_sequencer_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_sequencer_proto_goTypes,
		DependencyIndexes: file_sequencer_proto_depIdxs,
		MessageInfos:      file_sequencer_proto_msgTypes,
	}.Build()
	File_sequencer_proto = out.File
	file_sequencer_proto_rawDesc = nil
	file_sequencer_proto_goTypes = nil
	file_sequencer_proto_depIdxs = nil
}
