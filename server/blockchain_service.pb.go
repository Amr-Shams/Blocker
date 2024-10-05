// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: server/blockchain_service.proto

package server

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

type Block struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Timestamp    int64          `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Hash         []byte         `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	Transactions []*Transaction `protobuf:"bytes,3,rep,name=transactions,proto3" json:"transactions,omitempty"`
	PrevHash     []byte         `protobuf:"bytes,4,opt,name=prev_hash,json=prevHash,proto3" json:"prev_hash,omitempty"`
	Nonce        int32          `protobuf:"varint,5,opt,name=nonce,proto3" json:"nonce,omitempty"`
}

func (x *Block) Reset() {
	*x = Block{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{0}
}

func (x *Block) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Block) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *Block) GetTransactions() []*Transaction {
	if x != nil {
		return x.Transactions
	}
	return nil
}

func (x *Block) GetPrevHash() []byte {
	if x != nil {
		return x.PrevHash
	}
	return nil
}

func (x *Block) GetNonce() int32 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

type Transaction struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   []byte      `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Vin  []*TXInput  `protobuf:"bytes,2,rep,name=vin,proto3" json:"vin,omitempty"`
	Vout []*TXOutput `protobuf:"bytes,3,rep,name=vout,proto3" json:"vout,omitempty"`
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{1}
}

func (x *Transaction) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *Transaction) GetVin() []*TXInput {
	if x != nil {
		return x.Vin
	}
	return nil
}

func (x *Transaction) GetVout() []*TXOutput {
	if x != nil {
		return x.Vout
	}
	return nil
}

type TXInput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Txid      []byte `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	Vout      int32  `protobuf:"varint,2,opt,name=vout,proto3" json:"vout,omitempty"`
	Signature []byte `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	PubKey    []byte `protobuf:"bytes,4,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"`
}

func (x *TXInput) Reset() {
	*x = TXInput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TXInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TXInput) ProtoMessage() {}

func (x *TXInput) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TXInput.ProtoReflect.Descriptor instead.
func (*TXInput) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{2}
}

func (x *TXInput) GetTxid() []byte {
	if x != nil {
		return x.Txid
	}
	return nil
}

func (x *TXInput) GetVout() int32 {
	if x != nil {
		return x.Vout
	}
	return 0
}

func (x *TXInput) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *TXInput) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

type TXOutput struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Value      int32  `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
	PubKeyHash []byte `protobuf:"bytes,2,opt,name=pub_key_hash,json=pubKeyHash,proto3" json:"pub_key_hash,omitempty"`
}

func (x *TXOutput) Reset() {
	*x = TXOutput{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TXOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TXOutput) ProtoMessage() {}

func (x *TXOutput) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TXOutput.ProtoReflect.Descriptor instead.
func (*TXOutput) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{3}
}

func (x *TXOutput) GetValue() int32 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *TXOutput) GetPubKeyHash() []byte {
	if x != nil {
		return x.PubKeyHash
	}
	return nil
}

type TXOutputs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Outputs []*TXOutput `protobuf:"bytes,1,rep,name=outputs,proto3" json:"outputs,omitempty"`
}

func (x *TXOutputs) Reset() {
	*x = TXOutputs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TXOutputs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TXOutputs) ProtoMessage() {}

func (x *TXOutputs) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TXOutputs.ProtoReflect.Descriptor instead.
func (*TXOutputs) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{4}
}

func (x *TXOutputs) GetOutputs() []*TXOutput {
	if x != nil {
		return x.Outputs
	}
	return nil
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Blocks  []*Block `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
	Status  string   `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Success bool     `protobuf:"varint,3,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{5}
}

func (x *Response) GetBlocks() []*Block {
	if x != nil {
		return x.Blocks
	}
	return nil
}

func (x *Response) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *Response) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId  int32   `protobuf:"varint,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Address *string `protobuf:"bytes,2,opt,name=address,proto3,oneof" json:"address,omitempty"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{6}
}

func (x *Request) GetNodeId() int32 {
	if x != nil {
		return x.NodeId
	}
	return 0
}

func (x *Request) GetAddress() string {
	if x != nil && x.Address != nil {
		return *x.Address
	}
	return ""
}

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{7}
}

type HelloResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId  int32    `protobuf:"varint,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Address string   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Status  string   `protobuf:"bytes,3,opt,name=status,proto3" json:"status,omitempty"`
	Version string   `protobuf:"bytes,4,opt,name=version,proto3" json:"version,omitempty"`
	Blocks  []*Block `protobuf:"bytes,5,rep,name=blocks,proto3" json:"blocks,omitempty"`
	Peers   []string `protobuf:"bytes,6,rep,name=peers,proto3" json:"peers,omitempty"`
}

func (x *HelloResponse) Reset() {
	*x = HelloResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_server_blockchain_service_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HelloResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HelloResponse) ProtoMessage() {}

func (x *HelloResponse) ProtoReflect() protoreflect.Message {
	mi := &file_server_blockchain_service_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HelloResponse.ProtoReflect.Descriptor instead.
func (*HelloResponse) Descriptor() ([]byte, []int) {
	return file_server_blockchain_service_proto_rawDescGZIP(), []int{8}
}

func (x *HelloResponse) GetNodeId() int32 {
	if x != nil {
		return x.NodeId
	}
	return 0
}

func (x *HelloResponse) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *HelloResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *HelloResponse) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *HelloResponse) GetBlocks() []*Block {
	if x != nil {
		return x.Blocks
	}
	return nil
}

func (x *HelloResponse) GetPeers() []string {
	if x != nil {
		return x.Peers
	}
	return nil
}

var File_server_blockchain_service_proto protoreflect.FileDescriptor

var file_server_blockchain_service_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x06, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0xa5, 0x01, 0x0a, 0x05, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x37, 0x0a, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x0c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1b,
	0x0a, 0x09, 0x70, 0x72, 0x65, 0x76, 0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x08, 0x70, 0x72, 0x65, 0x76, 0x48, 0x61, 0x73, 0x68, 0x12, 0x14, 0x0a, 0x05, 0x6e,
	0x6f, 0x6e, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6e, 0x6f, 0x6e, 0x63,
	0x65, 0x22, 0x66, 0x0a, 0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x21, 0x0a, 0x03, 0x76, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x58, 0x49, 0x6e, 0x70, 0x75, 0x74, 0x52, 0x03,
	0x76, 0x69, 0x6e, 0x12, 0x24, 0x0a, 0x04, 0x76, 0x6f, 0x75, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x58, 0x4f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x52, 0x04, 0x76, 0x6f, 0x75, 0x74, 0x22, 0x68, 0x0a, 0x07, 0x54, 0x58, 0x49,
	0x6e, 0x70, 0x75, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x78, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x74, 0x78, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x6f, 0x75, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x76, 0x6f, 0x75, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x09, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x70, 0x75,
	0x62, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x75, 0x62,
	0x4b, 0x65, 0x79, 0x22, 0x42, 0x0a, 0x08, 0x54, 0x58, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x20, 0x0a, 0x0c, 0x70, 0x75, 0x62, 0x5f, 0x6b, 0x65, 0x79,
	0x5f, 0x68, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0a, 0x70, 0x75, 0x62,
	0x4b, 0x65, 0x79, 0x48, 0x61, 0x73, 0x68, 0x22, 0x37, 0x0a, 0x09, 0x54, 0x58, 0x4f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x73, 0x12, 0x2a, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x54,
	0x58, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x52, 0x07, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x73,
	0x22, 0x63, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x06,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x06, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x4d, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x1d, 0x0a, 0x07, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x07, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x88, 0x01, 0x01, 0x42, 0x0a, 0x0a, 0x08, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0xb1, 0x01,
	0x0a, 0x0d, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x06, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x52, 0x06, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x70,
	0x65, 0x65, 0x72, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x70, 0x65, 0x65, 0x72,
	0x73, 0x32, 0x94, 0x02, 0x0a, 0x11, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x30, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x12, 0x0d, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x37, 0x0a, 0x0e, 0x42, 0x72, 0x6f,
	0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x13, 0x2e, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x1a, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x30, 0x0a, 0x0b, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x0f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x31, 0x0a, 0x08, 0x41, 0x64, 0x64, 0x42, 0x6c, 0x6f, 0x63, 0x6b,
	0x12, 0x13, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f,
	0x12, 0x0f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x15, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x48, 0x65, 0x6c, 0x6c, 0x6f,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x25, 0x5a, 0x23, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x41, 0x6d, 0x72, 0x2d, 0x53, 0x68, 0x61, 0x6d, 0x73,
	0x2f, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x72, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_server_blockchain_service_proto_rawDescOnce sync.Once
	file_server_blockchain_service_proto_rawDescData = file_server_blockchain_service_proto_rawDesc
)

func file_server_blockchain_service_proto_rawDescGZIP() []byte {
	file_server_blockchain_service_proto_rawDescOnce.Do(func() {
		file_server_blockchain_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_server_blockchain_service_proto_rawDescData)
	})
	return file_server_blockchain_service_proto_rawDescData
}

var file_server_blockchain_service_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_server_blockchain_service_proto_goTypes = []any{
	(*Block)(nil),         // 0: server.Block
	(*Transaction)(nil),   // 1: server.Transaction
	(*TXInput)(nil),       // 2: server.TXInput
	(*TXOutput)(nil),      // 3: server.TXOutput
	(*TXOutputs)(nil),     // 4: server.TXOutputs
	(*Response)(nil),      // 5: server.Response
	(*Request)(nil),       // 6: server.Request
	(*Empty)(nil),         // 7: server.Empty
	(*HelloResponse)(nil), // 8: server.HelloResponse
}
var file_server_blockchain_service_proto_depIdxs = []int32{
	1,  // 0: server.Block.transactions:type_name -> server.Transaction
	2,  // 1: server.Transaction.vin:type_name -> server.TXInput
	3,  // 2: server.Transaction.vout:type_name -> server.TXOutput
	3,  // 3: server.TXOutputs.outputs:type_name -> server.TXOutput
	0,  // 4: server.Response.blocks:type_name -> server.Block
	0,  // 5: server.HelloResponse.blocks:type_name -> server.Block
	7,  // 6: server.BlockchainService.GetBlockChain:input_type -> server.Empty
	1,  // 7: server.BlockchainService.BroadcastBlock:input_type -> server.Transaction
	6,  // 8: server.BlockchainService.CheckStatus:input_type -> server.Request
	1,  // 9: server.BlockchainService.AddBlock:input_type -> server.Transaction
	6,  // 10: server.BlockchainService.Hello:input_type -> server.Request
	5,  // 11: server.BlockchainService.GetBlockChain:output_type -> server.Response
	5,  // 12: server.BlockchainService.BroadcastBlock:output_type -> server.Response
	5,  // 13: server.BlockchainService.CheckStatus:output_type -> server.Response
	5,  // 14: server.BlockchainService.AddBlock:output_type -> server.Response
	8,  // 15: server.BlockchainService.Hello:output_type -> server.HelloResponse
	11, // [11:16] is the sub-list for method output_type
	6,  // [6:11] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_server_blockchain_service_proto_init() }
func file_server_blockchain_service_proto_init() {
	if File_server_blockchain_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_server_blockchain_service_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*Block); i {
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
		file_server_blockchain_service_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*Transaction); i {
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
		file_server_blockchain_service_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*TXInput); i {
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
		file_server_blockchain_service_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*TXOutput); i {
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
		file_server_blockchain_service_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*TXOutputs); i {
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
		file_server_blockchain_service_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*Response); i {
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
		file_server_blockchain_service_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*Request); i {
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
		file_server_blockchain_service_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*Empty); i {
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
		file_server_blockchain_service_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*HelloResponse); i {
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
	file_server_blockchain_service_proto_msgTypes[6].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_server_blockchain_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_server_blockchain_service_proto_goTypes,
		DependencyIndexes: file_server_blockchain_service_proto_depIdxs,
		MessageInfos:      file_server_blockchain_service_proto_msgTypes,
	}.Build()
	File_server_blockchain_service_proto = out.File
	file_server_blockchain_service_proto_rawDesc = nil
	file_server_blockchain_service_proto_goTypes = nil
	file_server_blockchain_service_proto_depIdxs = nil
}