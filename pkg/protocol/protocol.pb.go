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

type MessageType int32

const (
	MessageType_PEER_IDENTITY    MessageType = 0
	MessageType_IMAGE_GENERATION MessageType = 1
)

// Enum value maps for MessageType.
var (
	MessageType_name = map[int32]string{
		0: "PEER_IDENTITY",
		1: "IMAGE_GENERATION",
	}
	MessageType_value = map[string]int32{
		"PEER_IDENTITY":    0,
		"IMAGE_GENERATION": 1,
	}
)

func (x MessageType) Enum() *MessageType {
	p := new(MessageType)
	*p = x
	return p
}

func (x MessageType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MessageType) Descriptor() protoreflect.EnumDescriptor {
	return file_protocol_proto_enumTypes[0].Descriptor()
}

func (MessageType) Type() protoreflect.EnumType {
	return &file_protocol_proto_enumTypes[0]
}

func (x MessageType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MessageType.Descriptor instead.
func (MessageType) EnumDescriptor() ([]byte, []int) {
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
	Receiver      string `protobuf:"bytes,5,opt,name=receiver,proto3" json:"receiver,omitempty"`
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

func (x *MessageHeader) GetReceiver() string {
	if x != nil {
		return x.Receiver
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

	Header        *MessageHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Type          MessageType    `protobuf:"varint,2,opt,name=type,proto3,enum=protocol.MessageType" json:"type,omitempty"`
	Body          []byte         `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
	ResultCode    int32          `protobuf:"varint,4,opt,name=resultCode,proto3" json:"resultCode,omitempty"`
	ResultMessage string         `protobuf:"bytes,5,opt,name=resultMessage,proto3" json:"resultMessage,omitempty"`
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

func (x *Message) GetType() MessageType {
	if x != nil {
		return x.Type
	}
	return MessageType_PEER_IDENTITY
}

func (x *Message) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *Message) GetResultCode() int32 {
	if x != nil {
		return x.ResultCode
	}
	return 0
}

func (x *Message) GetResultMessage() string {
	if x != nil {
		return x.ResultMessage
	}
	return ""
}

type PeerIdentityBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//
	//	*PeerIdentityBody_Req
	//	*PeerIdentityBody_Res
	Data isPeerIdentityBody_Data `protobuf_oneof:"data"`
}

func (x *PeerIdentityBody) Reset() {
	*x = PeerIdentityBody{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerIdentityBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerIdentityBody) ProtoMessage() {}

func (x *PeerIdentityBody) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PeerIdentityBody.ProtoReflect.Descriptor instead.
func (*PeerIdentityBody) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{2}
}

func (m *PeerIdentityBody) GetData() isPeerIdentityBody_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *PeerIdentityBody) GetReq() *PeerIdentityRequest {
	if x, ok := x.GetData().(*PeerIdentityBody_Req); ok {
		return x.Req
	}
	return nil
}

func (x *PeerIdentityBody) GetRes() *PeerIdentityResponse {
	if x, ok := x.GetData().(*PeerIdentityBody_Res); ok {
		return x.Res
	}
	return nil
}

type isPeerIdentityBody_Data interface {
	isPeerIdentityBody_Data()
}

type PeerIdentityBody_Req struct {
	Req *PeerIdentityRequest `protobuf:"bytes,1,opt,name=req,proto3,oneof"`
}

type PeerIdentityBody_Res struct {
	Res *PeerIdentityResponse `protobuf:"bytes,2,opt,name=res,proto3,oneof"`
}

func (*PeerIdentityBody_Req) isPeerIdentityBody_Data() {}

func (*PeerIdentityBody_Res) isPeerIdentityBody_Data() {}

type PeerIdentityRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId string `protobuf:"bytes,1,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
}

func (x *PeerIdentityRequest) Reset() {
	*x = PeerIdentityRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerIdentityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerIdentityRequest) ProtoMessage() {}

func (x *PeerIdentityRequest) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use PeerIdentityRequest.ProtoReflect.Descriptor instead.
func (*PeerIdentityRequest) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{3}
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
		mi := &file_protocol_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PeerIdentityResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PeerIdentityResponse) ProtoMessage() {}

func (x *PeerIdentityResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[4]
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
	return file_protocol_proto_rawDescGZIP(), []int{4}
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

type ImageGenerationBody struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Data:
	//
	//	*ImageGenerationBody_Req
	//	*ImageGenerationBody_Res
	Data isImageGenerationBody_Data `protobuf_oneof:"data"`
}

func (x *ImageGenerationBody) Reset() {
	*x = ImageGenerationBody{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageGenerationBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageGenerationBody) ProtoMessage() {}

func (x *ImageGenerationBody) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageGenerationBody.ProtoReflect.Descriptor instead.
func (*ImageGenerationBody) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{5}
}

func (m *ImageGenerationBody) GetData() isImageGenerationBody_Data {
	if m != nil {
		return m.Data
	}
	return nil
}

func (x *ImageGenerationBody) GetReq() *ImageGenerationRequest {
	if x, ok := x.GetData().(*ImageGenerationBody_Req); ok {
		return x.Req
	}
	return nil
}

func (x *ImageGenerationBody) GetRes() *ImageGenerationResponse {
	if x, ok := x.GetData().(*ImageGenerationBody_Res); ok {
		return x.Res
	}
	return nil
}

type isImageGenerationBody_Data interface {
	isImageGenerationBody_Data()
}

type ImageGenerationBody_Req struct {
	Req *ImageGenerationRequest `protobuf:"bytes,1,opt,name=req,proto3,oneof"`
}

type ImageGenerationBody_Res struct {
	Res *ImageGenerationResponse `protobuf:"bytes,2,opt,name=res,proto3,oneof"`
}

func (*ImageGenerationBody_Req) isImageGenerationBody_Data() {}

func (*ImageGenerationBody_Res) isImageGenerationBody_Data() {}

type ImageGenerationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId     string   `protobuf:"bytes,1,opt,name=nodeId,proto3" json:"nodeId,omitempty"`
	Model      string   `protobuf:"bytes,2,opt,name=model,proto3" json:"model,omitempty"`
	PromptWord []string `protobuf:"bytes,3,rep,name=promptWord,proto3" json:"promptWord,omitempty"`
}

func (x *ImageGenerationRequest) Reset() {
	*x = ImageGenerationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageGenerationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageGenerationRequest) ProtoMessage() {}

func (x *ImageGenerationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageGenerationRequest.ProtoReflect.Descriptor instead.
func (*ImageGenerationRequest) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{6}
}

func (x *ImageGenerationRequest) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *ImageGenerationRequest) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *ImageGenerationRequest) GetPromptWord() []string {
	if x != nil {
		return x.PromptWord
	}
	return nil
}

type ImageGenerationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IpfsNode  string `protobuf:"bytes,1,opt,name=ipfsNode,proto3" json:"ipfsNode,omitempty"`
	Cid       string `protobuf:"bytes,2,opt,name=cid,proto3" json:"cid,omitempty"`
	ImageName string `protobuf:"bytes,3,opt,name=imageName,proto3" json:"imageName,omitempty"`
}

func (x *ImageGenerationResponse) Reset() {
	*x = ImageGenerationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protocol_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageGenerationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageGenerationResponse) ProtoMessage() {}

func (x *ImageGenerationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protocol_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageGenerationResponse.ProtoReflect.Descriptor instead.
func (*ImageGenerationResponse) Descriptor() ([]byte, []int) {
	return file_protocol_proto_rawDescGZIP(), []int{7}
}

func (x *ImageGenerationResponse) GetIpfsNode() string {
	if x != nil {
		return x.IpfsNode
	}
	return ""
}

func (x *ImageGenerationResponse) GetCid() string {
	if x != nil {
		return x.Cid
	}
	return ""
}

func (x *ImageGenerationResponse) GetImageName() string {
	if x != nil {
		return x.ImageName
	}
	return ""
}

var File_protocol_proto protoreflect.FileDescriptor

var file_protocol_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x22, 0xcb, 0x01, 0x0a, 0x0d, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x0d,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x63, 0x65,
	0x69, 0x76, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x63, 0x65,
	0x69, 0x76, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x50, 0x75, 0x62, 0x4b,
	0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x6e, 0x6f, 0x64, 0x65, 0x50, 0x75,
	0x62, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x67, 0x6e, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x73, 0x69, 0x67, 0x6e, 0x22, 0xbf, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x2f, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x29, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x4d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x62, 0x6f, 0x64, 0x79, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x43, 0x6f,
	0x64, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x43, 0x6f, 0x64, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x72, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x81, 0x01, 0x0a, 0x10, 0x50,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x42, 0x6f, 0x64, 0x79, 0x12,
	0x31, 0x0a, 0x03, 0x72, 0x65, 0x71, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x03, 0x72,
	0x65, 0x71, 0x12, 0x32, 0x0a, 0x03, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x50, 0x65, 0x65, 0x72, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48,
	0x00, 0x52, 0x03, 0x72, 0x65, 0x73, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x2d,
	0x0a, 0x13, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x22, 0xc2, 0x01,
	0x0a, 0x14, 0x50, 0x65, 0x65, 0x72, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x22, 0x0a, 0x0c, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b,
	0x65, 0x79, 0x12, 0x20, 0x0a, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x41, 0x64, 0x64, 0x72,
	0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0b, 0x6c, 0x69, 0x73, 0x74, 0x65, 0x6e, 0x41,
	0x64, 0x64, 0x72, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x73, 0x22, 0x8a, 0x01, 0x0a, 0x13, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x47, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x6f, 0x64, 0x79, 0x12, 0x34, 0x0a, 0x03, 0x72, 0x65,
	0x71, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x03, 0x72, 0x65, 0x71,
	0x12, 0x35, 0x0a, 0x03, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x21, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x47, 0x65,
	0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x48, 0x00, 0x52, 0x03, 0x72, 0x65, 0x73, 0x42, 0x06, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22,
	0x66, 0x0a, 0x16, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6e, 0x6f, 0x64,
	0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6d, 0x70,
	0x74, 0x57, 0x6f, 0x72, 0x64, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x72, 0x6f,
	0x6d, 0x70, 0x74, 0x57, 0x6f, 0x72, 0x64, 0x22, 0x65, 0x0a, 0x17, 0x49, 0x6d, 0x61, 0x67, 0x65,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x70, 0x66, 0x73, 0x4e, 0x6f, 0x64, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x70, 0x66, 0x73, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x63, 0x69, 0x64,
	0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x2a, 0x36,
	0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x11, 0x0a,
	0x0d, 0x50, 0x45, 0x45, 0x52, 0x5f, 0x49, 0x44, 0x45, 0x4e, 0x54, 0x49, 0x54, 0x59, 0x10, 0x00,
	0x12, 0x14, 0x0a, 0x10, 0x49, 0x4d, 0x41, 0x47, 0x45, 0x5f, 0x47, 0x45, 0x4e, 0x45, 0x52, 0x41,
	0x54, 0x49, 0x4f, 0x4e, 0x10, 0x01, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x2e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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
var file_protocol_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_protocol_proto_goTypes = []interface{}{
	(MessageType)(0),                // 0: protocol.MessageType
	(*MessageHeader)(nil),           // 1: protocol.MessageHeader
	(*Message)(nil),                 // 2: protocol.Message
	(*PeerIdentityBody)(nil),        // 3: protocol.PeerIdentityBody
	(*PeerIdentityRequest)(nil),     // 4: protocol.PeerIdentityRequest
	(*PeerIdentityResponse)(nil),    // 5: protocol.PeerIdentityResponse
	(*ImageGenerationBody)(nil),     // 6: protocol.ImageGenerationBody
	(*ImageGenerationRequest)(nil),  // 7: protocol.ImageGenerationRequest
	(*ImageGenerationResponse)(nil), // 8: protocol.ImageGenerationResponse
}
var file_protocol_proto_depIdxs = []int32{
	1, // 0: protocol.Message.header:type_name -> protocol.MessageHeader
	0, // 1: protocol.Message.type:type_name -> protocol.MessageType
	4, // 2: protocol.PeerIdentityBody.req:type_name -> protocol.PeerIdentityRequest
	5, // 3: protocol.PeerIdentityBody.res:type_name -> protocol.PeerIdentityResponse
	7, // 4: protocol.ImageGenerationBody.req:type_name -> protocol.ImageGenerationRequest
	8, // 5: protocol.ImageGenerationBody.res:type_name -> protocol.ImageGenerationResponse
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
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
			switch v := v.(*PeerIdentityBody); i {
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
		file_protocol_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
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
		file_protocol_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImageGenerationBody); i {
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
		file_protocol_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImageGenerationRequest); i {
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
		file_protocol_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImageGenerationResponse); i {
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
	file_protocol_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*PeerIdentityBody_Req)(nil),
		(*PeerIdentityBody_Res)(nil),
	}
	file_protocol_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*ImageGenerationBody_Req)(nil),
		(*ImageGenerationBody_Res)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protocol_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   8,
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
