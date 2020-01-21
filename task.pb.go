// Code generated by protoc-gen-go. DO NOT EDIT.
// source: task.proto

package notifierhub

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type TorrentTask struct {
	Uuid                 string   `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Path                 string   `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Size                 int64    `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Type                 string   `protobuf:"bytes,4,opt,name=type,proto3" json:"type,omitempty"`
	Rest                 string   `protobuf:"bytes,5,opt,name=rest,proto3" json:"rest,omitempty"`
	Hash                 string   `protobuf:"bytes,6,opt,name=hash,proto3" json:"hash,omitempty"`
	StartTS              int64    `protobuf:"varint,7,opt,name=startTS,proto3" json:"startTS,omitempty"`
	FinishTS             int64    `protobuf:"varint,8,opt,name=finishTS,proto3" json:"finishTS,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TorrentTask) Reset()         { *m = TorrentTask{} }
func (m *TorrentTask) String() string { return proto.CompactTextString(m) }
func (*TorrentTask) ProtoMessage()    {}
func (*TorrentTask) Descriptor() ([]byte, []int) {
	return fileDescriptor_ce5d8dd45b4a91ff, []int{0}
}

func (m *TorrentTask) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TorrentTask.Unmarshal(m, b)
}
func (m *TorrentTask) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TorrentTask.Marshal(b, m, deterministic)
}
func (m *TorrentTask) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TorrentTask.Merge(m, src)
}
func (m *TorrentTask) XXX_Size() int {
	return xxx_messageInfo_TorrentTask.Size(m)
}
func (m *TorrentTask) XXX_DiscardUnknown() {
	xxx_messageInfo_TorrentTask.DiscardUnknown(m)
}

var xxx_messageInfo_TorrentTask proto.InternalMessageInfo

func (m *TorrentTask) GetUuid() string {
	if m != nil {
		return m.Uuid
	}
	return ""
}

func (m *TorrentTask) GetPath() string {
	if m != nil {
		return m.Path
	}
	return ""
}

func (m *TorrentTask) GetSize() int64 {
	if m != nil {
		return m.Size
	}
	return 0
}

func (m *TorrentTask) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *TorrentTask) GetRest() string {
	if m != nil {
		return m.Rest
	}
	return ""
}

func (m *TorrentTask) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *TorrentTask) GetStartTS() int64 {
	if m != nil {
		return m.StartTS
	}
	return 0
}

func (m *TorrentTask) GetFinishTS() int64 {
	if m != nil {
		return m.FinishTS
	}
	return 0
}

func init() {
	proto.RegisterType((*TorrentTask)(nil), "notifierhub.TorrentTask")
}

func init() { proto.RegisterFile("task.proto", fileDescriptor_ce5d8dd45b4a91ff) }

var fileDescriptor_ce5d8dd45b4a91ff = []byte{
	// 172 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0xcf, 0x31, 0x0e, 0x83, 0x20,
	0x14, 0xc6, 0xf1, 0x50, 0xad, 0x5a, 0xdc, 0x98, 0x5e, 0x3a, 0x99, 0x4e, 0x4e, 0x5d, 0x7a, 0x13,
	0xe5, 0x02, 0x98, 0x62, 0x20, 0x26, 0x62, 0x78, 0x8f, 0xa1, 0xbd, 0x5a, 0x2f, 0xd7, 0x3c, 0x48,
	0xbb, 0xfd, 0xbf, 0x1f, 0x84, 0x04, 0x29, 0xc9, 0xe0, 0x76, 0x3f, 0x62, 0xa0, 0xa0, 0xfa, 0x3d,
	0x90, 0x5f, 0xbd, 0x8d, 0x2e, 0x2d, 0xb7, 0x8f, 0x90, 0xbd, 0x0e, 0x31, 0xda, 0x9d, 0xb4, 0xc1,
	0x4d, 0x29, 0x59, 0xa7, 0xe4, 0x9f, 0x20, 0x06, 0x31, 0x5e, 0xa6, 0xdc, 0x6c, 0x87, 0x21, 0x07,
	0xa7, 0x62, 0xdc, 0x6c, 0xe8, 0xdf, 0x16, 0xaa, 0x41, 0x8c, 0xd5, 0x94, 0x9b, 0x8d, 0x5e, 0x87,
	0x85, 0xba, 0xdc, 0xe3, 0x66, 0x8b, 0x16, 0x09, 0xce, 0xc5, 0xb8, 0xd9, 0x9c, 0x41, 0x07, 0x4d,
	0x31, 0x6e, 0x05, 0xb2, 0x45, 0x32, 0x91, 0xf4, 0x0c, 0x6d, 0x7e, 0xf2, 0x37, 0xd5, 0x55, 0x76,
	0xab, 0xdf, 0x3d, 0x3a, 0x3d, 0x43, 0x97, 0x8f, 0xfe, 0x7b, 0x69, 0xf2, 0x8f, 0x1e, 0xdf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xd3, 0xb1, 0x82, 0x1c, 0xdf, 0x00, 0x00, 0x00,
}
