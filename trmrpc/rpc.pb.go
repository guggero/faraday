// Code generated by protoc-gen-go. DO NOT EDIT.
// source: rpc.proto

package trmrpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
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

type CloseRecommendationsRequest struct {
	//
	//The minimum amount of time in seconds that a channel should have been
	//monitored by lnd to be eligible for close. This value is in place to
	//protect against closing of newer channels.
	MinimumMonitored int64 `protobuf:"varint,1,opt,name=minimum_monitored,json=minimumMonitored,proto3" json:"minimum_monitored,omitempty"`
	//
	//The number of inter-quartile ranges a value needs to be beneath the lower
	//quartile/ above the upper quartile to be considered a lower/upper outlier.
	//Lower values will be more aggressive in recommending channel closes, and
	//upper values will be more conservative. Recommended values are 1.5 for
	//aggressive recommendations and 3 for conservative recommendations.
	OutlierMultiplier float32 `protobuf:"fixed32,2,opt,name=outlier_multiplier,json=outlierMultiplier,proto3" json:"outlier_multiplier,omitempty"`
	//
	//Threshold contains the threshold value that is used to recommend channels
	//for closure.
	//
	// Types that are valid to be assigned to Threshold:
	//	*CloseRecommendationsRequest_UptimeThreshold
	Threshold            isCloseRecommendationsRequest_Threshold `protobuf_oneof:"threshold"`
	XXX_NoUnkeyedLiteral struct{}                                `json:"-"`
	XXX_unrecognized     []byte                                  `json:"-"`
	XXX_sizecache        int32                                   `json:"-"`
}

func (m *CloseRecommendationsRequest) Reset()         { *m = CloseRecommendationsRequest{} }
func (m *CloseRecommendationsRequest) String() string { return proto.CompactTextString(m) }
func (*CloseRecommendationsRequest) ProtoMessage()    {}
func (*CloseRecommendationsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_77a6da22d6a3feb1, []int{0}
}

func (m *CloseRecommendationsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CloseRecommendationsRequest.Unmarshal(m, b)
}
func (m *CloseRecommendationsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CloseRecommendationsRequest.Marshal(b, m, deterministic)
}
func (m *CloseRecommendationsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CloseRecommendationsRequest.Merge(m, src)
}
func (m *CloseRecommendationsRequest) XXX_Size() int {
	return xxx_messageInfo_CloseRecommendationsRequest.Size(m)
}
func (m *CloseRecommendationsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CloseRecommendationsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CloseRecommendationsRequest proto.InternalMessageInfo

func (m *CloseRecommendationsRequest) GetMinimumMonitored() int64 {
	if m != nil {
		return m.MinimumMonitored
	}
	return 0
}

func (m *CloseRecommendationsRequest) GetOutlierMultiplier() float32 {
	if m != nil {
		return m.OutlierMultiplier
	}
	return 0
}

type isCloseRecommendationsRequest_Threshold interface {
	isCloseRecommendationsRequest_Threshold()
}

type CloseRecommendationsRequest_UptimeThreshold struct {
	UptimeThreshold float32 `protobuf:"fixed32,3,opt,name=uptime_threshold,json=uptimeThreshold,proto3,oneof"`
}

func (*CloseRecommendationsRequest_UptimeThreshold) isCloseRecommendationsRequest_Threshold() {}

func (m *CloseRecommendationsRequest) GetThreshold() isCloseRecommendationsRequest_Threshold {
	if m != nil {
		return m.Threshold
	}
	return nil
}

func (m *CloseRecommendationsRequest) GetUptimeThreshold() float32 {
	if x, ok := m.GetThreshold().(*CloseRecommendationsRequest_UptimeThreshold); ok {
		return x.UptimeThreshold
	}
	return 0
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*CloseRecommendationsRequest) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*CloseRecommendationsRequest_UptimeThreshold)(nil),
	}
}

type CloseRecommendationsResponse struct {
	//
	//The total number of channels, before filtering out channels that are
	//not eligible for close recommendations.
	TotalChannels int32 `protobuf:"varint,1,opt,name=total_channels,json=totalChannels,proto3" json:"total_channels,omitempty"`
	//
	//The number of channels that were considered for close recommendations.
	ConsideredChannels int32 `protobuf:"varint,2,opt,name=considered_channels,json=consideredChannels,proto3" json:"considered_channels,omitempty"`
	//
	//A map of channels to close recommendations, based out whether they are
	//outliers in the uptime dataset. The absence of a channel in this set
	//implies that it was not considered for close because it did not meet
	//the criteria for close (it is private, or has not been monitored for
	//long enough to make a decision).
	OutlierRecommendations []*Recommendation `protobuf:"bytes,3,rep,name=outlier_recommendations,json=outlierRecommendations,proto3" json:"outlier_recommendations,omitempty"`
	//
	//A set of channel close recommendations, based out whether they are
	//beneath the threshold provided in the request. The absence of a channel
	//in this set implies that it was not considered for close because it
	//did not meet the criteria for close (it is private, or has not been
	//monitored for long enough to make a decision).
	ThresholdRecommendations []*Recommendation `protobuf:"bytes,4,rep,name=threshold_recommendations,json=thresholdRecommendations,proto3" json:"threshold_recommendations,omitempty"`
	XXX_NoUnkeyedLiteral     struct{}          `json:"-"`
	XXX_unrecognized         []byte            `json:"-"`
	XXX_sizecache            int32             `json:"-"`
}

func (m *CloseRecommendationsResponse) Reset()         { *m = CloseRecommendationsResponse{} }
func (m *CloseRecommendationsResponse) String() string { return proto.CompactTextString(m) }
func (*CloseRecommendationsResponse) ProtoMessage()    {}
func (*CloseRecommendationsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_77a6da22d6a3feb1, []int{1}
}

func (m *CloseRecommendationsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CloseRecommendationsResponse.Unmarshal(m, b)
}
func (m *CloseRecommendationsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CloseRecommendationsResponse.Marshal(b, m, deterministic)
}
func (m *CloseRecommendationsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CloseRecommendationsResponse.Merge(m, src)
}
func (m *CloseRecommendationsResponse) XXX_Size() int {
	return xxx_messageInfo_CloseRecommendationsResponse.Size(m)
}
func (m *CloseRecommendationsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CloseRecommendationsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CloseRecommendationsResponse proto.InternalMessageInfo

func (m *CloseRecommendationsResponse) GetTotalChannels() int32 {
	if m != nil {
		return m.TotalChannels
	}
	return 0
}

func (m *CloseRecommendationsResponse) GetConsideredChannels() int32 {
	if m != nil {
		return m.ConsideredChannels
	}
	return 0
}

func (m *CloseRecommendationsResponse) GetOutlierRecommendations() []*Recommendation {
	if m != nil {
		return m.OutlierRecommendations
	}
	return nil
}

func (m *CloseRecommendationsResponse) GetThresholdRecommendations() []*Recommendation {
	if m != nil {
		return m.ThresholdRecommendations
	}
	return nil
}

type Recommendation struct {
	//
	//The channel point [funding txid: outpoint] of the channel being considered
	//for close.
	ChanPoint string `protobuf:"bytes,1,opt,name=chan_point,json=chanPoint,proto3" json:"chan_point,omitempty"`
	// The value of the metric that close recommendations were based on.
	Value float32 `protobuf:"fixed32,2,opt,name=value,proto3" json:"value,omitempty"`
	// A boolean indicating whether we recommend closing the channel.
	RecommendClose       bool     `protobuf:"varint,3,opt,name=recommend_close,json=recommendClose,proto3" json:"recommend_close,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Recommendation) Reset()         { *m = Recommendation{} }
func (m *Recommendation) String() string { return proto.CompactTextString(m) }
func (*Recommendation) ProtoMessage()    {}
func (*Recommendation) Descriptor() ([]byte, []int) {
	return fileDescriptor_77a6da22d6a3feb1, []int{2}
}

func (m *Recommendation) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Recommendation.Unmarshal(m, b)
}
func (m *Recommendation) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Recommendation.Marshal(b, m, deterministic)
}
func (m *Recommendation) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Recommendation.Merge(m, src)
}
func (m *Recommendation) XXX_Size() int {
	return xxx_messageInfo_Recommendation.Size(m)
}
func (m *Recommendation) XXX_DiscardUnknown() {
	xxx_messageInfo_Recommendation.DiscardUnknown(m)
}

var xxx_messageInfo_Recommendation proto.InternalMessageInfo

func (m *Recommendation) GetChanPoint() string {
	if m != nil {
		return m.ChanPoint
	}
	return ""
}

func (m *Recommendation) GetValue() float32 {
	if m != nil {
		return m.Value
	}
	return 0
}

func (m *Recommendation) GetRecommendClose() bool {
	if m != nil {
		return m.RecommendClose
	}
	return false
}

func init() {
	proto.RegisterType((*CloseRecommendationsRequest)(nil), "trmrpc.CloseRecommendationsRequest")
	proto.RegisterType((*CloseRecommendationsResponse)(nil), "trmrpc.CloseRecommendationsResponse")
	proto.RegisterType((*Recommendation)(nil), "trmrpc.Recommendation")
}

func init() { proto.RegisterFile("rpc.proto", fileDescriptor_77a6da22d6a3feb1) }

var fileDescriptor_77a6da22d6a3feb1 = []byte{
	// 387 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x52, 0xed, 0x6e, 0xd3, 0x40,
	0x10, 0xc4, 0x36, 0xad, 0xf0, 0x56, 0xa4, 0xe9, 0x51, 0x15, 0x53, 0x8a, 0x14, 0x05, 0x10, 0x91,
	0x2a, 0x12, 0xa9, 0xbc, 0x01, 0xfd, 0xc3, 0x9f, 0x0a, 0x74, 0xc9, 0x7f, 0xeb, 0xb0, 0x57, 0xc9,
	0x49, 0x77, 0xb7, 0xc7, 0xdd, 0x39, 0x4f, 0xc3, 0x5b, 0xf0, 0x82, 0xc8, 0x9f, 0x51, 0xd2, 0x28,
	0xff, 0xec, 0x99, 0xd9, 0xbd, 0xd9, 0x9d, 0x85, 0xd4, 0xd9, 0x62, 0x6e, 0x1d, 0x05, 0x62, 0xe7,
	0xc1, 0x69, 0x67, 0x8b, 0xdb, 0xbb, 0x35, 0xd1, 0x5a, 0xe1, 0x42, 0x58, 0xb9, 0x10, 0xc6, 0x50,
	0x10, 0x41, 0x92, 0xf1, 0xad, 0x6a, 0xfa, 0x2f, 0x82, 0xf7, 0x8f, 0x8a, 0x3c, 0x72, 0x2c, 0x48,
	0x6b, 0x34, 0x65, 0x4b, 0x73, 0xfc, 0x53, 0xa1, 0x0f, 0xec, 0x1e, 0xae, 0xb4, 0x34, 0x52, 0x57,
	0x3a, 0xd7, 0x64, 0x64, 0x20, 0x87, 0x65, 0x16, 0x4d, 0xa2, 0x59, 0xc2, 0xc7, 0x1d, 0xf1, 0xd4,
	0xe3, 0xec, 0x2b, 0x30, 0xaa, 0x82, 0x92, 0xe8, 0x72, 0x5d, 0xa9, 0x20, 0x6d, 0xfd, 0x99, 0xc5,
	0x93, 0x68, 0x16, 0xf3, 0xab, 0x8e, 0x79, 0x1a, 0x08, 0x76, 0x0f, 0xe3, 0xca, 0x06, 0xa9, 0x31,
	0x0f, 0x1b, 0x87, 0x7e, 0x43, 0xaa, 0xcc, 0x92, 0x5a, 0xfc, 0xe3, 0x05, 0xbf, 0x6c, 0x99, 0x55,
	0x4f, 0x7c, 0xbf, 0x80, 0x74, 0x50, 0x4d, 0xff, 0xc6, 0x70, 0x77, 0xdc, 0xb5, 0xb7, 0x64, 0x3c,
	0xb2, 0xcf, 0x30, 0x0a, 0x14, 0x84, 0xca, 0x8b, 0x8d, 0x30, 0x06, 0x95, 0x6f, 0x3c, 0x9f, 0xf1,
	0xd7, 0x0d, 0xfa, 0xd8, 0x81, 0x6c, 0x01, 0x6f, 0x0a, 0x32, 0x5e, 0x96, 0xe8, 0xb0, 0xdc, 0x69,
	0xe3, 0x46, 0xcb, 0x76, 0xd4, 0x50, 0xf0, 0x13, 0xde, 0xf6, 0x13, 0xba, 0xfd, 0xa7, 0xb3, 0x64,
	0x92, 0xcc, 0x2e, 0x1e, 0x6e, 0xe6, 0xed, 0xda, 0xe7, 0xfb, 0xce, 0xf8, 0x4d, 0x57, 0x76, 0x60,
	0x98, 0x2d, 0xe1, 0xdd, 0x30, 0xd6, 0xb3, 0x96, 0x2f, 0x4f, 0xb6, 0xcc, 0x86, 0xc2, 0x83, 0xa6,
	0x53, 0x03, 0xa3, 0x7d, 0x88, 0x7d, 0x00, 0xa8, 0xa7, 0xcb, 0x2d, 0x49, 0x13, 0x9a, 0x5d, 0xa4,
	0x3c, 0xad, 0x91, 0x5f, 0x35, 0xc0, 0xae, 0xe1, 0x6c, 0x2b, 0x54, 0x85, 0x5d, 0x56, 0xed, 0x0f,
	0xfb, 0x02, 0x97, 0x83, 0xa3, 0xbc, 0xa8, 0xd7, 0xdd, 0xc4, 0xf3, 0x8a, 0x8f, 0x06, 0xb8, 0x09,
	0xe1, 0xa1, 0x82, 0xf1, 0x0a, 0x9d, 0x96, 0x46, 0x04, 0x72, 0x4b, 0x74, 0x5b, 0x74, 0x4c, 0xc0,
	0xf5, 0xb1, 0x84, 0xd8, 0xc7, 0x7e, 0x9a, 0x13, 0x57, 0x77, 0xfb, 0xe9, 0xb4, 0xa8, 0x0d, 0xf9,
	0xf7, 0x79, 0x73, 0xc2, 0xdf, 0xfe, 0x07, 0x00, 0x00, 0xff, 0xff, 0x3f, 0x08, 0x22, 0xe1, 0xf5,
	0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TerminatorServerClient is the client API for TerminatorServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TerminatorServerClient interface {
	CloseRecommendations(ctx context.Context, in *CloseRecommendationsRequest, opts ...grpc.CallOption) (*CloseRecommendationsResponse, error)
}

type terminatorServerClient struct {
	cc *grpc.ClientConn
}

func NewTerminatorServerClient(cc *grpc.ClientConn) TerminatorServerClient {
	return &terminatorServerClient{cc}
}

func (c *terminatorServerClient) CloseRecommendations(ctx context.Context, in *CloseRecommendationsRequest, opts ...grpc.CallOption) (*CloseRecommendationsResponse, error) {
	out := new(CloseRecommendationsResponse)
	err := c.cc.Invoke(ctx, "/trmrpc.TerminatorServer/CloseRecommendations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TerminatorServerServer is the server API for TerminatorServer service.
type TerminatorServerServer interface {
	CloseRecommendations(context.Context, *CloseRecommendationsRequest) (*CloseRecommendationsResponse, error)
}

func RegisterTerminatorServerServer(s *grpc.Server, srv TerminatorServerServer) {
	s.RegisterService(&_TerminatorServer_serviceDesc, srv)
}

func _TerminatorServer_CloseRecommendations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseRecommendationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TerminatorServerServer).CloseRecommendations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trmrpc.TerminatorServer/CloseRecommendations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TerminatorServerServer).CloseRecommendations(ctx, req.(*CloseRecommendationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TerminatorServer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "trmrpc.TerminatorServer",
	HandlerType: (*TerminatorServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CloseRecommendations",
			Handler:    _TerminatorServer_CloseRecommendations_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rpc.proto",
}