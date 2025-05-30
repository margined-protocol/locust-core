// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dydxprotocol/clob/liquidations.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	types "github.com/margined-protocol/locust-core/pkg/proto/dydx/subaccounts/types"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// PerpetualLiquidationInfo holds information about a liquidation that occurred
// for a position held by a subaccount.
// Note this proto is defined to make it easier to hash
// the metadata of a liquidation, and is never written to state.
type PerpetualLiquidationInfo struct {
	// The id of the subaccount that got liquidated/deleveraged or was deleveraged
	// onto.
	SubaccountId types.SubaccountId `protobuf:"bytes,1,opt,name=subaccount_id,json=subaccountId,proto3" json:"subaccount_id"`
	// The id of the perpetual involved.
	PerpetualId uint32 `protobuf:"varint,2,opt,name=perpetual_id,json=perpetualId,proto3" json:"perpetual_id,omitempty"`
}

func (m *PerpetualLiquidationInfo) Reset()         { *m = PerpetualLiquidationInfo{} }
func (m *PerpetualLiquidationInfo) String() string { return proto.CompactTextString(m) }
func (*PerpetualLiquidationInfo) ProtoMessage()    {}
func (*PerpetualLiquidationInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_e91628d349879778, []int{0}
}
func (m *PerpetualLiquidationInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PerpetualLiquidationInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PerpetualLiquidationInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PerpetualLiquidationInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PerpetualLiquidationInfo.Merge(m, src)
}
func (m *PerpetualLiquidationInfo) XXX_Size() int {
	return m.Size()
}
func (m *PerpetualLiquidationInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_PerpetualLiquidationInfo.DiscardUnknown(m)
}

var xxx_messageInfo_PerpetualLiquidationInfo proto.InternalMessageInfo

func (m *PerpetualLiquidationInfo) GetSubaccountId() types.SubaccountId {
	if m != nil {
		return m.SubaccountId
	}
	return types.SubaccountId{}
}

func (m *PerpetualLiquidationInfo) GetPerpetualId() uint32 {
	if m != nil {
		return m.PerpetualId
	}
	return 0
}

// SubaccountLiquidationInfo holds liquidation information per-subaccount in the
// current block.
type SubaccountLiquidationInfo struct {
	// An unsorted list of unique perpetual IDs that the subaccount has previously
	// liquidated.
	PerpetualsLiquidated []uint32 `protobuf:"varint,1,rep,packed,name=perpetuals_liquidated,json=perpetualsLiquidated,proto3" json:"perpetuals_liquidated,omitempty"`
	// The notional value (in quote quantums, determined by the oracle price) of
	// all positions liquidated for this subaccount.
	NotionalLiquidated uint64 `protobuf:"varint,2,opt,name=notional_liquidated,json=notionalLiquidated,proto3" json:"notional_liquidated,omitempty"`
	// The amount of funds that the insurance fund has lost
	// covering this subaccount.
	QuantumsInsuranceLost uint64 `protobuf:"varint,3,opt,name=quantums_insurance_lost,json=quantumsInsuranceLost,proto3" json:"quantums_insurance_lost,omitempty"`
}

func (m *SubaccountLiquidationInfo) Reset()         { *m = SubaccountLiquidationInfo{} }
func (m *SubaccountLiquidationInfo) String() string { return proto.CompactTextString(m) }
func (*SubaccountLiquidationInfo) ProtoMessage()    {}
func (*SubaccountLiquidationInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_e91628d349879778, []int{1}
}
func (m *SubaccountLiquidationInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SubaccountLiquidationInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SubaccountLiquidationInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SubaccountLiquidationInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubaccountLiquidationInfo.Merge(m, src)
}
func (m *SubaccountLiquidationInfo) XXX_Size() int {
	return m.Size()
}
func (m *SubaccountLiquidationInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SubaccountLiquidationInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SubaccountLiquidationInfo proto.InternalMessageInfo

func (m *SubaccountLiquidationInfo) GetPerpetualsLiquidated() []uint32 {
	if m != nil {
		return m.PerpetualsLiquidated
	}
	return nil
}

func (m *SubaccountLiquidationInfo) GetNotionalLiquidated() uint64 {
	if m != nil {
		return m.NotionalLiquidated
	}
	return 0
}

func (m *SubaccountLiquidationInfo) GetQuantumsInsuranceLost() uint64 {
	if m != nil {
		return m.QuantumsInsuranceLost
	}
	return 0
}

// SubaccountOpenPositionInfo holds information about open positions for a
// perpetual.
type SubaccountOpenPositionInfo struct {
	// The id of the perpetual.
	PerpetualId uint32 `protobuf:"varint,1,opt,name=perpetual_id,json=perpetualId,proto3" json:"perpetual_id,omitempty"`
	// The ids of the subaccounts with long positions.
	SubaccountsWithLongPosition []types.SubaccountId `protobuf:"bytes,2,rep,name=subaccounts_with_long_position,json=subaccountsWithLongPosition,proto3" json:"subaccounts_with_long_position"`
	// The ids of the subaccounts with short positions.
	SubaccountsWithShortPosition []types.SubaccountId `protobuf:"bytes,3,rep,name=subaccounts_with_short_position,json=subaccountsWithShortPosition,proto3" json:"subaccounts_with_short_position"`
}

func (m *SubaccountOpenPositionInfo) Reset()         { *m = SubaccountOpenPositionInfo{} }
func (m *SubaccountOpenPositionInfo) String() string { return proto.CompactTextString(m) }
func (*SubaccountOpenPositionInfo) ProtoMessage()    {}
func (*SubaccountOpenPositionInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_e91628d349879778, []int{2}
}
func (m *SubaccountOpenPositionInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *SubaccountOpenPositionInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_SubaccountOpenPositionInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *SubaccountOpenPositionInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SubaccountOpenPositionInfo.Merge(m, src)
}
func (m *SubaccountOpenPositionInfo) XXX_Size() int {
	return m.Size()
}
func (m *SubaccountOpenPositionInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_SubaccountOpenPositionInfo.DiscardUnknown(m)
}

var xxx_messageInfo_SubaccountOpenPositionInfo proto.InternalMessageInfo

func (m *SubaccountOpenPositionInfo) GetPerpetualId() uint32 {
	if m != nil {
		return m.PerpetualId
	}
	return 0
}

func (m *SubaccountOpenPositionInfo) GetSubaccountsWithLongPosition() []types.SubaccountId {
	if m != nil {
		return m.SubaccountsWithLongPosition
	}
	return nil
}

func (m *SubaccountOpenPositionInfo) GetSubaccountsWithShortPosition() []types.SubaccountId {
	if m != nil {
		return m.SubaccountsWithShortPosition
	}
	return nil
}

func init() {
	proto.RegisterType((*PerpetualLiquidationInfo)(nil), "dydxprotocol.clob.PerpetualLiquidationInfo")
	proto.RegisterType((*SubaccountLiquidationInfo)(nil), "dydxprotocol.clob.SubaccountLiquidationInfo")
	proto.RegisterType((*SubaccountOpenPositionInfo)(nil), "dydxprotocol.clob.SubaccountOpenPositionInfo")
}

func init() {
	proto.RegisterFile("dydxprotocol/clob/liquidations.proto", fileDescriptor_e91628d349879778)
}

var fileDescriptor_e91628d349879778 = []byte{
	// 415 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0xcd, 0xae, 0xd2, 0x40,
	0x18, 0xed, 0xd0, 0x1b, 0x17, 0x73, 0x2f, 0x0b, 0x2b, 0xc4, 0x8a, 0xa6, 0x20, 0x31, 0x06, 0x17,
	0xb6, 0x89, 0x18, 0x1e, 0x80, 0x5d, 0x93, 0x26, 0x62, 0x59, 0x98, 0xb8, 0x69, 0xfa, 0x67, 0x3b,
	0x49, 0x99, 0xaf, 0x74, 0x66, 0x14, 0xde, 0x82, 0xb7, 0xf0, 0x19, 0x7c, 0x03, 0x96, 0x2c, 0x5d,
	0x19, 0x03, 0x2f, 0x62, 0x5a, 0x6c, 0x3b, 0xc0, 0x4a, 0x77, 0x5f, 0xce, 0x39, 0xdf, 0x39, 0x67,
	0xa6, 0x1d, 0xfc, 0x2a, 0xda, 0x46, 0x9b, 0xbc, 0x00, 0x0e, 0x21, 0x64, 0x56, 0x98, 0x41, 0x60,
	0x65, 0x64, 0x2d, 0x48, 0xe4, 0x73, 0x02, 0x94, 0x99, 0x15, 0xa5, 0x3d, 0x96, 0x55, 0x66, 0xa9,
	0x1a, 0xf4, 0x12, 0x48, 0xa0, 0x82, 0xac, 0x72, 0x3a, 0x0b, 0x07, 0x6f, 0x2e, 0xec, 0x98, 0x08,
	0xfc, 0x30, 0x04, 0x41, 0x39, 0x93, 0xe6, 0xb3, 0x74, 0xbc, 0x43, 0x58, 0x5f, 0xc4, 0x45, 0x1e,
	0x73, 0xe1, 0x67, 0x4e, 0x9b, 0x69, 0xd3, 0x2f, 0xa0, 0x7d, 0xc4, 0xdd, 0x76, 0xc1, 0x23, 0x91,
	0x8e, 0x46, 0x68, 0x72, 0xff, 0xee, 0xb5, 0x79, 0x51, 0x44, 0xf2, 0x37, 0x97, 0xcd, 0x6c, 0x47,
	0xf3, 0xbb, 0xfd, 0xaf, 0xa1, 0xe2, 0x3e, 0x30, 0x09, 0xd3, 0x5e, 0xe2, 0x87, 0xbc, 0x8e, 0x2b,
	0x1d, 0x3b, 0x23, 0x34, 0xe9, 0xba, 0xf7, 0x0d, 0x66, 0x47, 0xe3, 0x1f, 0x08, 0x3f, 0x6b, 0x7d,
	0xae, 0x3b, 0x4d, 0x71, 0xbf, 0x11, 0x33, 0xaf, 0xbe, 0xa5, 0xb8, 0xec, 0xa6, 0x4e, 0xba, 0x6e,
	0xaf, 0x25, 0x9d, 0x86, 0xd3, 0x2c, 0xfc, 0x84, 0x42, 0x69, 0xe1, 0x67, 0xf2, 0x4a, 0x19, 0x7e,
	0xe7, 0x6a, 0x35, 0x25, 0x2d, 0xcc, 0xf0, 0xd3, 0xb5, 0xf0, 0x29, 0x17, 0x2b, 0xe6, 0x11, 0xca,
	0x44, 0xe1, 0xd3, 0x30, 0xf6, 0x32, 0x60, 0x5c, 0x57, 0xab, 0xa5, 0x7e, 0x4d, 0xdb, 0x35, 0xeb,
	0x00, 0xe3, 0xe3, 0xef, 0x1d, 0x3c, 0x68, 0xbb, 0x7f, 0xc8, 0x63, 0xba, 0x00, 0x46, 0x9a, 0xf2,
	0xd7, 0xa7, 0x47, 0x37, 0xa7, 0xd7, 0xd6, 0xd8, 0x90, 0x2e, 0xd4, 0xfb, 0x46, 0x78, 0xea, 0x65,
	0x40, 0x13, 0x2f, 0xff, 0x6b, 0xa4, 0x77, 0x46, 0xea, 0x3f, 0x7f, 0x84, 0xe7, 0x12, 0xff, 0x89,
	0xf0, 0xd4, 0x01, 0x9a, 0xd4, 0xcd, 0x34, 0x86, 0x87, 0x37, 0x91, 0x2c, 0x85, 0x82, 0xb7, 0x99,
	0xea, 0x7f, 0x64, 0xbe, 0xb8, 0xca, 0x5c, 0x96, 0x96, 0x75, 0xe8, 0x7c, 0xb1, 0x3f, 0x1a, 0xe8,
	0x70, 0x34, 0xd0, 0xef, 0xa3, 0x81, 0x76, 0x27, 0x43, 0x39, 0x9c, 0x0c, 0xe5, 0xe7, 0xc9, 0x50,
	0x3e, 0xcf, 0x12, 0xc2, 0x53, 0x11, 0x98, 0x21, 0xac, 0xac, 0x8b, 0x1f, 0xf9, 0xeb, 0xfb, 0xb7,
	0x61, 0xea, 0x13, 0x6a, 0x35, 0xc8, 0xe6, 0xfc, 0x56, 0xf8, 0x36, 0x8f, 0x59, 0xf0, 0xa8, 0x82,
	0xa7, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x46, 0x48, 0x06, 0xe2, 0x4d, 0x03, 0x00, 0x00,
}

func (m *PerpetualLiquidationInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PerpetualLiquidationInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PerpetualLiquidationInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PerpetualId != 0 {
		i = encodeVarintLiquidations(dAtA, i, uint64(m.PerpetualId))
		i--
		dAtA[i] = 0x10
	}
	{
		size, err := m.SubaccountId.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintLiquidations(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *SubaccountLiquidationInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SubaccountLiquidationInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SubaccountLiquidationInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.QuantumsInsuranceLost != 0 {
		i = encodeVarintLiquidations(dAtA, i, uint64(m.QuantumsInsuranceLost))
		i--
		dAtA[i] = 0x18
	}
	if m.NotionalLiquidated != 0 {
		i = encodeVarintLiquidations(dAtA, i, uint64(m.NotionalLiquidated))
		i--
		dAtA[i] = 0x10
	}
	if len(m.PerpetualsLiquidated) > 0 {
		dAtA3 := make([]byte, len(m.PerpetualsLiquidated)*10)
		var j2 int
		for _, num := range m.PerpetualsLiquidated {
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		i -= j2
		copy(dAtA[i:], dAtA3[:j2])
		i = encodeVarintLiquidations(dAtA, i, uint64(j2))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *SubaccountOpenPositionInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *SubaccountOpenPositionInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *SubaccountOpenPositionInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.SubaccountsWithShortPosition) > 0 {
		for iNdEx := len(m.SubaccountsWithShortPosition) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SubaccountsWithShortPosition[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintLiquidations(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.SubaccountsWithLongPosition) > 0 {
		for iNdEx := len(m.SubaccountsWithLongPosition) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.SubaccountsWithLongPosition[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintLiquidations(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.PerpetualId != 0 {
		i = encodeVarintLiquidations(dAtA, i, uint64(m.PerpetualId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintLiquidations(dAtA []byte, offset int, v uint64) int {
	offset -= sovLiquidations(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *PerpetualLiquidationInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.SubaccountId.Size()
	n += 1 + l + sovLiquidations(uint64(l))
	if m.PerpetualId != 0 {
		n += 1 + sovLiquidations(uint64(m.PerpetualId))
	}
	return n
}

func (m *SubaccountLiquidationInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.PerpetualsLiquidated) > 0 {
		l = 0
		for _, e := range m.PerpetualsLiquidated {
			l += sovLiquidations(uint64(e))
		}
		n += 1 + sovLiquidations(uint64(l)) + l
	}
	if m.NotionalLiquidated != 0 {
		n += 1 + sovLiquidations(uint64(m.NotionalLiquidated))
	}
	if m.QuantumsInsuranceLost != 0 {
		n += 1 + sovLiquidations(uint64(m.QuantumsInsuranceLost))
	}
	return n
}

func (m *SubaccountOpenPositionInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.PerpetualId != 0 {
		n += 1 + sovLiquidations(uint64(m.PerpetualId))
	}
	if len(m.SubaccountsWithLongPosition) > 0 {
		for _, e := range m.SubaccountsWithLongPosition {
			l = e.Size()
			n += 1 + l + sovLiquidations(uint64(l))
		}
	}
	if len(m.SubaccountsWithShortPosition) > 0 {
		for _, e := range m.SubaccountsWithShortPosition {
			l = e.Size()
			n += 1 + l + sovLiquidations(uint64(l))
		}
	}
	return n
}

func sovLiquidations(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozLiquidations(x uint64) (n int) {
	return sovLiquidations(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *PerpetualLiquidationInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLiquidations
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: PerpetualLiquidationInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PerpetualLiquidationInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SubaccountId", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLiquidations
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLiquidations
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SubaccountId.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerpetualId", wireType)
			}
			m.PerpetualId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerpetualId |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLiquidations(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLiquidations
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SubaccountLiquidationInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLiquidations
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SubaccountLiquidationInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SubaccountLiquidationInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v uint32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowLiquidations
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.PerpetualsLiquidated = append(m.PerpetualsLiquidated, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowLiquidations
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthLiquidations
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthLiquidations
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.PerpetualsLiquidated) == 0 {
					m.PerpetualsLiquidated = make([]uint32, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowLiquidations
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint32(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.PerpetualsLiquidated = append(m.PerpetualsLiquidated, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field PerpetualsLiquidated", wireType)
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NotionalLiquidated", wireType)
			}
			m.NotionalLiquidated = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NotionalLiquidated |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field QuantumsInsuranceLost", wireType)
			}
			m.QuantumsInsuranceLost = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.QuantumsInsuranceLost |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipLiquidations(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLiquidations
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *SubaccountOpenPositionInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowLiquidations
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: SubaccountOpenPositionInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: SubaccountOpenPositionInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PerpetualId", wireType)
			}
			m.PerpetualId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PerpetualId |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SubaccountsWithLongPosition", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLiquidations
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLiquidations
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SubaccountsWithLongPosition = append(m.SubaccountsWithLongPosition, types.SubaccountId{})
			if err := m.SubaccountsWithLongPosition[len(m.SubaccountsWithLongPosition)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SubaccountsWithShortPosition", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthLiquidations
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthLiquidations
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SubaccountsWithShortPosition = append(m.SubaccountsWithShortPosition, types.SubaccountId{})
			if err := m.SubaccountsWithShortPosition[len(m.SubaccountsWithShortPosition)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipLiquidations(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthLiquidations
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipLiquidations(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowLiquidations
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowLiquidations
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthLiquidations
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupLiquidations
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthLiquidations
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthLiquidations        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowLiquidations          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupLiquidations = fmt.Errorf("proto: unexpected end of group")
)
