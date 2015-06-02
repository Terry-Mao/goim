// Code generated by protoc-gen-gogo.
// source: router.proto
// DO NOT EDIT!

/*
	Package main is a generated protocol buffer package.

	It is generated from these files:
		router.proto

	It has these top-level messages:
		PbRPCSubKey
		PbRPCSubRet
		PbRPCSetSubArg
*/
package main

import proto "github.com/gogo/protobuf/proto"
import math "math"

// discarding unused import gogoproto "gogo/protobuf/gogoproto/gogo.pb"

import io "io"
import fmt "fmt"
import github_com_gogo_protobuf_proto "github.com/gogo/protobuf/proto"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = math.Inf

type PbRPCSubKey struct {
	Key              string `protobuf:"bytes,1,opt,name=key" json:"key"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *PbRPCSubKey) Reset()         { *m = PbRPCSubKey{} }
func (m *PbRPCSubKey) String() string { return proto.CompactTextString(m) }
func (*PbRPCSubKey) ProtoMessage()    {}

func (m *PbRPCSubKey) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

type PbRPCSubRet struct {
	Ret              int64  `protobuf:"varint,1,opt,name=ret" json:"ret"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *PbRPCSubRet) Reset()         { *m = PbRPCSubRet{} }
func (m *PbRPCSubRet) String() string { return proto.CompactTextString(m) }
func (*PbRPCSubRet) ProtoMessage()    {}

func (m *PbRPCSubRet) GetRet() int64 {
	if m != nil {
		return m.Ret
	}
	return 0
}

type PbRPCSetSubArg struct {
	Subkey           string `protobuf:"bytes,1,opt,name=subkey" json:"subkey"`
	State            int32  `protobuf:"varint,2,opt,name=state" json:"state"`
	Server           int32  `protobuf:"varint,3,opt,name=server" json:"server"`
	XXX_unrecognized []byte `json:"-"`
}

func (m *PbRPCSetSubArg) Reset()         { *m = PbRPCSetSubArg{} }
func (m *PbRPCSetSubArg) String() string { return proto.CompactTextString(m) }
func (*PbRPCSetSubArg) ProtoMessage()    {}

func (m *PbRPCSetSubArg) GetSubkey() string {
	if m != nil {
		return m.Subkey
	}
	return ""
}

func (m *PbRPCSetSubArg) GetState() int32 {
	if m != nil {
		return m.State
	}
	return 0
}

func (m *PbRPCSetSubArg) GetServer() int32 {
	if m != nil {
		return m.Server
	}
	return 0
}

func init() {
}
func (m *PbRPCSubKey) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + int(stringLen)
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Key = string(data[index:postIndex])
			index = postIndex
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := github_com_gogo_protobuf_proto.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}

	return nil
}
func (m *PbRPCSubRet) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ret", wireType)
			}
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				m.Ret |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := github_com_gogo_protobuf_proto.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}

	return nil
}
func (m *PbRPCSetSubArg) Unmarshal(data []byte) error {
	l := len(data)
	index := 0
	for index < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if index >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[index]
			index++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Subkey", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			postIndex := index + int(stringLen)
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Subkey = string(data[index:postIndex])
			index = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field State", wireType)
			}
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				m.State |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Server", wireType)
			}
			for shift := uint(0); ; shift += 7 {
				if index >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[index]
				index++
				m.Server |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			var sizeOfWire int
			for {
				sizeOfWire++
				wire >>= 7
				if wire == 0 {
					break
				}
			}
			index -= sizeOfWire
			skippy, err := github_com_gogo_protobuf_proto.Skip(data[index:])
			if err != nil {
				return err
			}
			if (index + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, data[index:index+skippy]...)
			index += skippy
		}
	}

	return nil
}
func (m *PbRPCSubKey) Size() (n int) {
	var l int
	_ = l
	l = len(m.Key)
	n += 1 + l + sovRouter(uint64(l))
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *PbRPCSubRet) Size() (n int) {
	var l int
	_ = l
	n += 1 + sovRouter(uint64(m.Ret))
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func (m *PbRPCSetSubArg) Size() (n int) {
	var l int
	_ = l
	l = len(m.Subkey)
	n += 1 + l + sovRouter(uint64(l))
	n += 1 + sovRouter(uint64(m.State))
	n += 1 + sovRouter(uint64(m.Server))
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovRouter(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozRouter(x uint64) (n int) {
	return sovRouter(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PbRPCSubKey) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PbRPCSubKey) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0xa
	i++
	i = encodeVarintRouter(data, i, uint64(len(m.Key)))
	i += copy(data[i:], m.Key)
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *PbRPCSubRet) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PbRPCSubRet) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0x8
	i++
	i = encodeVarintRouter(data, i, uint64(m.Ret))
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func (m *PbRPCSetSubArg) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *PbRPCSetSubArg) MarshalTo(data []byte) (n int, err error) {
	var i int
	_ = i
	var l int
	_ = l
	data[i] = 0xa
	i++
	i = encodeVarintRouter(data, i, uint64(len(m.Subkey)))
	i += copy(data[i:], m.Subkey)
	data[i] = 0x10
	i++
	i = encodeVarintRouter(data, i, uint64(m.State))
	data[i] = 0x18
	i++
	i = encodeVarintRouter(data, i, uint64(m.Server))
	if m.XXX_unrecognized != nil {
		i += copy(data[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeFixed64Router(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Router(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintRouter(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}