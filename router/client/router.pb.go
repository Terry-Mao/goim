package main

import proto "github.com/golang/protobuf/proto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type PbRPCSubKey struct {
	Key string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
}

func (m *PbRPCSubKey) Reset()         { *m = PbRPCSubKey{} }
func (m *PbRPCSubKey) String() string { return proto.CompactTextString(m) }
func (*PbRPCSubKey) ProtoMessage()    {}

type PbRPCSubRet struct {
	Ret int64 `protobuf:"varint,1,opt,name=ret" json:"ret,omitempty"`
}

func (m *PbRPCSubRet) Reset()         { *m = PbRPCSubRet{} }
func (m *PbRPCSubRet) String() string { return proto.CompactTextString(m) }
func (*PbRPCSubRet) ProtoMessage()    {}

type PbRPCSetSubArg struct {
	Subkey string `protobuf:"bytes,1,opt,name=subkey" json:"subkey,omitempty"`
	State  int32  `protobuf:"varint,2,opt,name=state" json:"state,omitempty"`
	Server int32  `protobuf:"varint,3,opt,name=server" json:"server,omitempty"`
}

func (m *PbRPCSetSubArg) Reset()         { *m = PbRPCSetSubArg{} }
func (m *PbRPCSetSubArg) String() string { return proto.CompactTextString(m) }
func (*PbRPCSetSubArg) ProtoMessage()    {}

func init() {
}
