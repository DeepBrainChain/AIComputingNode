// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.6.1
// source: protocol.proto

package protocol

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

type MesasgeType int32

const (
	MesasgeType_PEER_IDENTITY_REQUEST  MesasgeType = 0
	MesasgeType_PEER_IDENTITY_RESPONSE MesasgeType = 1
)

// Enum value maps for MesasgeType.
var (
	MesasgeType_name = map[int32]string{
		0: "PEER_IDENTITY_REQUEST",
		1: "PEER_IDENTITY_RESPONSE",
	}
	MesasgeType_value = map[string]int32{
		"PEER_IDENTITY_REQUEST":  0,
		"PEER_IDENTITY_RESPONSE": 1,
	}
)

func (x MesasgeType) Enum() *MesasgeType {
	p := new(MesasgeType)
	*p = x
	return p
}

func (x MesasgeType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MesasgeType) Descriptor() protoreflect.EnumDescriptor {
	return file_protocol_proto_enumTypes[0].Descriptor()
}

func (MesasgeType) Type() protoreflect.EnumType {
	return &file_protocol_proto_enumTypes[0]
}

func (x MesasgeType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MesasgeType.Descriptor instead.
func (MesasgeType) EnumDescriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{0}
}

type MessageHeader struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ClientVersion string `protobuf:"bytes,1,opt,name=clientVersion,proto3" json:"clientVersion,omitempty"`
	Timestamp     int64  `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Id            string `protobuf:"bytes,3,opt,name=id,proto3" json:"id,omitempty"`
	NodeId        string `protobuf:"bytes,4,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	NodePubKey    []byte `protobuf:"bytes,6,opt,name=nodePubKey,proto3" json:"nodePubKey,omitempty"`
	Sign          []byte `protobuf:"bytes,7,opt,name=sign,proto3" json:"sign,omitempty"`
}

func (x *MessageHeader) Reset() {
	*x = MessageHeader{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageHeader) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageHeader) ProtoMessage() {}

func (x *MessageHeader) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageHeader.ProtoReflect.Descriptor instead.
func (*MessageHeader) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{0}
}

func (x *MessageHeader) GetClientVersion() string {
	if x != nil {
		return x.ClientVersion
	}
	return ""
}

func (x *MessageHeader) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *MessageHeader) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *MessageHeader) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *MessageHeader) GetNodePubKey() []byte {
	if x != nil {
		return x.NodePubKey
	}
	return nil
}

func (x *MessageHeader) GetSign() []byte {
	if x != nil {
		return x.Sign
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Header *MessageHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Type   MesasgeType    `protobuf:"varint,2,opt,name=type,proto3,enum=protocol.MesasgeType" json:"type,omitempty"`
	// Types that are assignable to Body:
	//
	//	*Message_PiReq
	//	*Message_PiRes
	Body isMessage_Body `protobuf_oneof:"body"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{1}
}

func (x *Message) GetHeader() *MessageHeader {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *Message) GetType() MesasgeType {
	if x != nil {
		return x.Type
	}
	return MesasgeType_PEER_IDENTITY_REQUEST
}

func (m *Message) GetBody() isMessage_Body {
	if m != nil {
		return m.Body
	}
	return nil
}

func (x *Message) GetPiReq() *PeerIdentityRequest {
	if x, ok := x.GetBody().(*Message_PiReq); ok {
		return x.PiReq
	}
	return nil
}

func (x *Message) GetPiRes() *PeerIdentityResponse {
	if x, ok := x.GetBody().(*Message_PiRes); ok {
		return x.PiRes
	}
	return nil
}

type isMessage_Body interface {
	isMessage_Body()
}

type Message_PiReq struct {
	PiReq *PeerIdentityRequest `protobuf:"bytes,3,opt,name=piReq,proto3,oneof"`
}

type Message_PiRes struct {
	PiRes *PeerIdentityResponse `protobuf:"bytes,4,opt,name=piRes,proto3,oneof"`
}

func (*Message_PiReq) isMessage_Body() {}

func (*Message_PiRes) isMessage_Body() {}

type PeerIdentityRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId string `protobuf:"bytes,1,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
}

func (x *PeerIdentityRequest) Reset() {
	*x = PeerIdentityRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerIdentityRequest) ProtoMessage() {}

func (x *PeerIdentityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerIdentityRequest.ProtoReflect.Descriptor instead.
func (*PeerIdentityRequest) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{2}
}

func (x *PeerIdentityRequest) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

type PeerIdentityResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProtocolVersion string   `protobuf:"bytes,5,opt,name=protocolVersion,proto3" json:"protocolVersion,omitempty"` // e.g. ipfs/1.0.0
	AgentVersion    string   `protobuf:"bytes,6,opt,name=agentVersion,proto3" json:"agentVersion,omitempty"`       // e.g. go-ipfs/0.1.0
	PublicKey       []byte   `protobuf:"bytes,1,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
	ListenAddrs     []string `protobuf:"bytes,2,rep,name=listenAddrs,proto3" json:"listenAddrs,omitempty"`
	Protocols       []string `protobuf:"bytes,3,rep,name=protocols,proto3" json:"protocols,omitempty"`
}

func (x *PeerIdentityResponse) Reset() {
	*x = PeerIdentityResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerIdentityResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerIdentityResponse) ProtoMessage() {}

func (x *PeerIdentityResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PeerIdentityResponse.ProtoReflect.Descriptor instead.
func (*PeerIdentityResponse) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{3}
}

func (x *PeerIdentityResponse) GetProtocolVersion() string {
	if x != nil {
		return x.ProtocolVersion
	}
	return ""
}

func (x *PeerIdentityResponse) GetAgentVersion() string {
	if x != nil {
		return x.AgentVersion
	}
	return ""
}

func (x *PeerIdentityResponse) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *PeerIdentityResponse) GetListenAddrs() []string {
	if x != nil {
		return x.ListenAddrs
	}
	return nil
}

func (x *PeerIdentityResponse) GetProtocols() []string {
	if x != nil {
		return x.Protocols
	}
	return nil
}

var File_protocol_proto protoreflect.FileDescriptor

var file_protocol_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0xaf, 0x01, 0x0a, 0x0d, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x0d,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x6e, 0x6f, 0x64, 0x65,
	0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x6e, 0x6f,
	0x64, 0x65, 0x50, 0x75, 0x62, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x67, 0x6e,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x73, 0x69, 0x67, 0x6e, 0x22, 0xdc, 0x01, 0x0a,
	0x07, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2f, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x63, 0x6f, 0x6c, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x2e, 0x4d, 0x65, 0x73, 0x61, 0x73, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x12, 0x35, 0x0a, 0x05, 0x70, 0x69, 0x52, 0x65, 0x71, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x50,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x48, 0x00, 0x52, 0x05, 0x70, 0x69, 0x52, 0x65, 0x71, 0x12, 0x36, 0x0a, 0x05, 0x70,
	0x69, 0x52, 0x65, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x05, 0x70, 0x69,
	0x52, 0x65, 0x73, 0x42, 0x06, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x22, 0x2d, 0x0a, 0x13, 0x50,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x22, 0xc2, 0x01, 0x0a, 0x14, 0x50,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56,
	0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a,
	0x0c, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0c, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12,
	0x20, 0x0a, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x41, 0x64, 0x64, 0x72, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x41, 0x64, 0x64, 0x72,
	0x73, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x73, 0x2a,
	0x44, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x61, 0x73, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19,
	0x0a, 0x15, 0x50, 0x45, 0x45, 0x52, 0x5f, 0x49, 0x44, 0x45, 0x4e, 0x54, 0x49, 0x54, 0x59, 0x5f,
	0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x10, 0x00, 0x12, 0x1a, 0x0a, 0x16, 0x50, 0x45, 0x45,
	0x52, 0x5f, 0x49, 0x44, 0x45, 0x4e, 0x54, 0x49, 0x54, 0x59, 0x5f, 0x52, 0x45, 0x53, 0x50, 0x4f,
	0x4e, 0x53, 0x45, 0x10, 0x01, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x2e, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protocol_proto_rawDescOnce sync.Once
	file_protocol_proto_rawDescData = file_protocol_proto_rawDesc
)

func file_protocol_proto_rawDescGZIP() []byte {
	file_protocol_proto_rawDescOnce.Do(func() {
		file_protocol_proto_rawDescData = protoimpl.X.CompressGZIP(file_protocol_proto_rawDescData)
	})
	return file_protocol_proto_rawDescData
}

var file_protocol_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_protocol_proto_goTypes = []interface{}{
	(MesasgeType)(0),             // 0: protocol.MesasgeType
	(*MessageHeader)(nil),        // 1: protocol.MessageHeader
	(*Message)(nil),              // 2: protocol.Message
	(*PeerIdentityRequest)(nil),  // 3: protocol.PeerIdentityRequest
	(*PeerIdentityResponse)(nil), // 4: protocol.PeerIdentityResponse
}
var file_protocol_proto_depIdxs = []int32{
	1, // 0: protocol.Message.header:type_name -> protocol.MessageHeader
	0, // 1: protocol.Message.type:type_name -> protocol.MesasgeType
	3, // 2: protocol.Message.piReq:type_name -> protocol.PeerIdentityRequest
	4, // 3: protocol.Message.piRes:type_name -> protocol.PeerIdentityResponse
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_protocol_proto_init() }
func file_protocol_proto_init() {
	if File_protocol_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protocol_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageHeader); i {
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
		file_protocol_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
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
		file_protocol_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerIdentityRequest); i {
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
		file_protocol_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PeerIdentityResponse); i {
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
	file_protocol_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Message_PiReq)(nil),
		(*Message_PiRes)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protocol_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protocol_proto_goTypes,
		DependencyIndexes: file_protocol_proto_depIdxs,
		EnumInfos:         file_protocol_proto_enumTypes,
		MessageInfos:      file_protocol_proto_msgTypes,
	}.Build()
	File_protocol_proto = out.File
	file_protocol_proto_rawDesc = nil
	file_protocol_proto_goTypes = nil
	file_protocol_proto_depIdxs = nil
}