// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.29.0
// 	protoc        v3.21.12
// source: pkg/echo/v1/echo.proto

package v1

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

type EchoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Query string `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
}

func (x *EchoRequest) Reset() {
	*x = EchoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EchoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EchoRequest) ProtoMessage() {}

func (x *EchoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EchoRequest.ProtoReflect.Descriptor instead.
func (*EchoRequest) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{0}
}

func (x *EchoRequest) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

type EchoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Query      string          `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
	Kubernetes *KubernetesInfo `protobuf:"bytes,2,opt,name=kubernetes,proto3" json:"kubernetes,omitempty"`
	Request    *RequestInfo    `protobuf:"bytes,3,opt,name=request,proto3" json:"request,omitempty"`
	Runtime    *RuntimeInfo    `protobuf:"bytes,4,opt,name=runtime,proto3" json:"runtime,omitempty"`
}

func (x *EchoResponse) Reset() {
	*x = EchoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EchoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EchoResponse) ProtoMessage() {}

func (x *EchoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EchoResponse.ProtoReflect.Descriptor instead.
func (*EchoResponse) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{1}
}

func (x *EchoResponse) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *EchoResponse) GetKubernetes() *KubernetesInfo {
	if x != nil {
		return x.Kubernetes
	}
	return nil
}

func (x *EchoResponse) GetRequest() *RequestInfo {
	if x != nil {
		return x.Request
	}
	return nil
}

func (x *EchoResponse) GetRuntime() *RuntimeInfo {
	if x != nil {
		return x.Runtime
	}
	return nil
}

type KubernetesInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hostname     string `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	PodName      string `protobuf:"bytes,2,opt,name=pod_name,json=podName,proto3" json:"pod_name,omitempty"`
	PodNamespace string `protobuf:"bytes,3,opt,name=pod_namespace,json=podNamespace,proto3" json:"pod_namespace,omitempty"`
	PodNode      string `protobuf:"bytes,4,opt,name=pod_node,json=podNode,proto3" json:"pod_node,omitempty"`
}

func (x *KubernetesInfo) Reset() {
	*x = KubernetesInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KubernetesInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KubernetesInfo) ProtoMessage() {}

func (x *KubernetesInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KubernetesInfo.ProtoReflect.Descriptor instead.
func (*KubernetesInfo) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{2}
}

func (x *KubernetesInfo) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *KubernetesInfo) GetPodName() string {
	if x != nil {
		return x.PodName
	}
	return ""
}

func (x *KubernetesInfo) GetPodNamespace() string {
	if x != nil {
		return x.PodNamespace
	}
	return ""
}

func (x *KubernetesInfo) GetPodNode() string {
	if x != nil {
		return x.PodNode
	}
	return ""
}

type RequestInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Protocol   string           `protobuf:"bytes,1,opt,name=protocol,proto3" json:"protocol,omitempty"`
	RemoteAddr string           `protobuf:"bytes,2,opt,name=remote_addr,json=remoteAddr,proto3" json:"remote_addr,omitempty"`
	Method     string           `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Uri        string           `protobuf:"bytes,4,opt,name=uri,proto3" json:"uri,omitempty"`
	ParsedUrl  *ParsedURL       `protobuf:"bytes,5,opt,name=parsed_url,json=parsedUrl,proto3" json:"parsed_url,omitempty"`
	Header     []*KeyMultivalue `protobuf:"bytes,6,rep,name=header,proto3" json:"header,omitempty"`
}

func (x *RequestInfo) Reset() {
	*x = RequestInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestInfo) ProtoMessage() {}

func (x *RequestInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestInfo.ProtoReflect.Descriptor instead.
func (*RequestInfo) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{3}
}

func (x *RequestInfo) GetProtocol() string {
	if x != nil {
		return x.Protocol
	}
	return ""
}

func (x *RequestInfo) GetRemoteAddr() string {
	if x != nil {
		return x.RemoteAddr
	}
	return ""
}

func (x *RequestInfo) GetMethod() string {
	if x != nil {
		return x.Method
	}
	return ""
}

func (x *RequestInfo) GetUri() string {
	if x != nil {
		return x.Uri
	}
	return ""
}

func (x *RequestInfo) GetParsedUrl() *ParsedURL {
	if x != nil {
		return x.ParsedUrl
	}
	return nil
}

func (x *RequestInfo) GetHeader() []*KeyMultivalue {
	if x != nil {
		return x.Header
	}
	return nil
}

type ParsedURL struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Scheme   string           `protobuf:"bytes,1,opt,name=scheme,proto3" json:"scheme,omitempty"`
	Host     string           `protobuf:"bytes,2,opt,name=host,proto3" json:"host,omitempty"`
	Path     string           `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	RawPath  string           `protobuf:"bytes,4,opt,name=raw_path,json=rawPath,proto3" json:"raw_path,omitempty"`
	RawQuery string           `protobuf:"bytes,5,opt,name=raw_query,json=rawQuery,proto3" json:"raw_query,omitempty"`
	Query    []*KeyMultivalue `protobuf:"bytes,6,rep,name=query,proto3" json:"query,omitempty"`
}

func (x *ParsedURL) Reset() {
	*x = ParsedURL{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParsedURL) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParsedURL) ProtoMessage() {}

func (x *ParsedURL) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParsedURL.ProtoReflect.Descriptor instead.
func (*ParsedURL) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{4}
}

func (x *ParsedURL) GetScheme() string {
	if x != nil {
		return x.Scheme
	}
	return ""
}

func (x *ParsedURL) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *ParsedURL) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *ParsedURL) GetRawPath() string {
	if x != nil {
		return x.RawPath
	}
	return ""
}

func (x *ParsedURL) GetRawQuery() string {
	if x != nil {
		return x.RawQuery
	}
	return ""
}

func (x *ParsedURL) GetQuery() []*KeyMultivalue {
	if x != nil {
		return x.Query
	}
	return nil
}

type KeyMultivalue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key    string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Values []string `protobuf:"bytes,2,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *KeyMultivalue) Reset() {
	*x = KeyMultivalue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyMultivalue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyMultivalue) ProtoMessage() {}

func (x *KeyMultivalue) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyMultivalue.ProtoReflect.Descriptor instead.
func (*KeyMultivalue) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{5}
}

func (x *KeyMultivalue) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KeyMultivalue) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

type RuntimeInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GoVersion     string `protobuf:"bytes,1,opt,name=go_version,json=goVersion,proto3" json:"go_version,omitempty"`
	GoArch        string `protobuf:"bytes,2,opt,name=go_arch,json=goArch,proto3" json:"go_arch,omitempty"`
	GoOs          string `protobuf:"bytes,3,opt,name=go_os,json=goOs,proto3" json:"go_os,omitempty"`
	NumCpus       int64  `protobuf:"varint,4,opt,name=num_cpus,json=numCpus,proto3" json:"num_cpus,omitempty"`
	NumGoroutines int64  `protobuf:"varint,5,opt,name=num_goroutines,json=numGoroutines,proto3" json:"num_goroutines,omitempty"`
	MainModule    string `protobuf:"bytes,6,opt,name=main_module,json=mainModule,proto3" json:"main_module,omitempty"`
	MainPath      string `protobuf:"bytes,7,opt,name=main_path,json=mainPath,proto3" json:"main_path,omitempty"`
	MainVersion   string `protobuf:"bytes,8,opt,name=main_version,json=mainVersion,proto3" json:"main_version,omitempty"`
}

func (x *RuntimeInfo) Reset() {
	*x = RuntimeInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RuntimeInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RuntimeInfo) ProtoMessage() {}

func (x *RuntimeInfo) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RuntimeInfo.ProtoReflect.Descriptor instead.
func (*RuntimeInfo) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{6}
}

func (x *RuntimeInfo) GetGoVersion() string {
	if x != nil {
		return x.GoVersion
	}
	return ""
}

func (x *RuntimeInfo) GetGoArch() string {
	if x != nil {
		return x.GoArch
	}
	return ""
}

func (x *RuntimeInfo) GetGoOs() string {
	if x != nil {
		return x.GoOs
	}
	return ""
}

func (x *RuntimeInfo) GetNumCpus() int64 {
	if x != nil {
		return x.NumCpus
	}
	return 0
}

func (x *RuntimeInfo) GetNumGoroutines() int64 {
	if x != nil {
		return x.NumGoroutines
	}
	return 0
}

func (x *RuntimeInfo) GetMainModule() string {
	if x != nil {
		return x.MainModule
	}
	return ""
}

func (x *RuntimeInfo) GetMainPath() string {
	if x != nil {
		return x.MainPath
	}
	return ""
}

func (x *RuntimeInfo) GetMainVersion() string {
	if x != nil {
		return x.MainVersion
	}
	return ""
}

// StatusRequest determines the error (if any) to return to the caller, representing
// on-wire status that the caller expects from the server.
type StatusRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// force_grpc_status forces a status response conforming to the standard
	// gRPC Core status codes: https://grpc.github.io/grpc/core/md_doc_statuscodes.html
	// Following those standard status codes, a value of zero (or unset) will
	// return OK.
	ForceGrpcStatus uint32 `protobuf:"varint,1,opt,name=force_grpc_status,json=forceGrpcStatus,proto3" json:"force_grpc_status,omitempty"`
	// message is an optional string returned alongside the code.
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *StatusRequest) Reset() {
	*x = StatusRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusRequest) ProtoMessage() {}

func (x *StatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusRequest.ProtoReflect.Descriptor instead.
func (*StatusRequest) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{7}
}

func (x *StatusRequest) GetForceGrpcStatus() uint32 {
	if x != nil {
		return x.ForceGrpcStatus
	}
	return 0
}

func (x *StatusRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

// StatusResponse is a (currently) empty gRPC message.
type StatusResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *StatusResponse) Reset() {
	*x = StatusResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_echo_v1_echo_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatusResponse) ProtoMessage() {}

func (x *StatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_echo_v1_echo_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatusResponse.ProtoReflect.Descriptor instead.
func (*StatusResponse) Descriptor() ([]byte, []int) {
	return file_pkg_echo_v1_echo_proto_rawDescGZIP(), []int{8}
}

var File_pkg_echo_v1_echo_proto protoreflect.FileDescriptor

var file_pkg_echo_v1_echo_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6b, 0x67, 0x2f, 0x65, 0x63, 0x68, 0x6f, 0x2f, 0x76, 0x31, 0x2f, 0x65, 0x63,
	0x68, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63,
	0x68, 0x6f, 0x2e, 0x76, 0x31, 0x22, 0x23, 0x0a, 0x0b, 0x45, 0x63, 0x68, 0x6f, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x22, 0xc9, 0x01, 0x0a, 0x0c, 0x45,
	0x63, 0x68, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x71,
	0x75, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x12, 0x3b, 0x0a, 0x0a, 0x6b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f,
	0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x49, 0x6e,
	0x66, 0x6f, 0x52, 0x0a, 0x6b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65, 0x74, 0x65, 0x73, 0x12, 0x32,
	0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x18, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x32, 0x0a, 0x07, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x07, 0x72,
	0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x22, 0x87, 0x01, 0x0a, 0x0e, 0x4b, 0x75, 0x62, 0x65, 0x72,
	0x6e, 0x65, 0x74, 0x65, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73,
	0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73,
	0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x6f, 0x64, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x23, 0x0a, 0x0d, 0x70, 0x6f, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x6f, 0x64, 0x5f, 0x6e, 0x6f, 0x64,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x6f, 0x64, 0x4e, 0x6f, 0x64, 0x65,
	0x22, 0xdf, 0x01, 0x0a, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x6e, 0x66, 0x6f,
	0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x12, 0x1f, 0x0a, 0x0b,
	0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x72, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72, 0x12, 0x16, 0x0a,
	0x06, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x69, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x69, 0x12, 0x35, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x73, 0x65,
	0x64, 0x5f, 0x75, 0x72, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x70, 0x6b,
	0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x72, 0x73, 0x65, 0x64,
	0x55, 0x52, 0x4c, 0x52, 0x09, 0x70, 0x61, 0x72, 0x73, 0x65, 0x64, 0x55, 0x72, 0x6c, 0x12, 0x32,
	0x0a, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x65, 0x79,
	0x4d, 0x75, 0x6c, 0x74, 0x69, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x22, 0xb5, 0x01, 0x0a, 0x09, 0x50, 0x61, 0x72, 0x73, 0x65, 0x64, 0x55, 0x52, 0x4c,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68,
	0x12, 0x19, 0x0a, 0x08, 0x72, 0x61, 0x77, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x72, 0x61, 0x77, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1b, 0x0a, 0x09, 0x72,
	0x61, 0x77, 0x5f, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08,
	0x72, 0x61, 0x77, 0x51, 0x75, 0x65, 0x72, 0x79, 0x12, 0x30, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72,
	0x79, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63,
	0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x65, 0x79, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x22, 0x39, 0x0a, 0x0d, 0x4b, 0x65,
	0x79, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a,
	0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0xfd, 0x01, 0x0a, 0x0b, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1d, 0x0a, 0x0a, 0x67, 0x6f, 0x5f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x67, 0x6f, 0x56, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x07, 0x67, 0x6f, 0x5f, 0x61, 0x72, 0x63, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x67, 0x6f, 0x41, 0x72, 0x63, 0x68, 0x12, 0x13, 0x0a,
	0x05, 0x67, 0x6f, 0x5f, 0x6f, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x67, 0x6f,
	0x4f, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x6e, 0x75, 0x6d, 0x5f, 0x63, 0x70, 0x75, 0x73, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x6e, 0x75, 0x6d, 0x43, 0x70, 0x75, 0x73, 0x12, 0x25, 0x0a,
	0x0e, 0x6e, 0x75, 0x6d, 0x5f, 0x67, 0x6f, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x65, 0x73, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0d, 0x6e, 0x75, 0x6d, 0x47, 0x6f, 0x72, 0x6f, 0x75, 0x74,
	0x69, 0x6e, 0x65, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x6d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6d, 0x61, 0x69, 0x6e, 0x4d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x70, 0x61,
	0x74, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x61, 0x69, 0x6e, 0x50, 0x61,
	0x74, 0x68, 0x12, 0x21, 0x0a, 0x0c, 0x6d, 0x61, 0x69, 0x6e, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6d, 0x61, 0x69, 0x6e, 0x56, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x55, 0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2a, 0x0a, 0x11, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x5f,
	0x67, 0x72, 0x70, 0x63, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x0f, 0x66, 0x6f, 0x72, 0x63, 0x65, 0x47, 0x72, 0x70, 0x63, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x10, 0x0a, 0x0e,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x8c,
	0x01, 0x0a, 0x06, 0x45, 0x63, 0x68, 0x6f, 0x65, 0x72, 0x12, 0x3d, 0x0a, 0x04, 0x45, 0x63, 0x68,
	0x6f, 0x12, 0x18, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e,
	0x45, 0x63, 0x68, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x70, 0x6b,
	0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x63, 0x68, 0x6f, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x43, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x1a, 0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b,
	0x2e, 0x70, 0x6b, 0x67, 0x2e, 0x65, 0x63, 0x68, 0x6f, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x27, 0x5a,
	0x25, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x69, 0x70, 0x74,
	0x61, 0x2f, 0x6e, 0x65, 0x74, 0x64, 0x65, 0x62, 0x75, 0x67, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x65,
	0x63, 0x68, 0x6f, 0x2f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_echo_v1_echo_proto_rawDescOnce sync.Once
	file_pkg_echo_v1_echo_proto_rawDescData = file_pkg_echo_v1_echo_proto_rawDesc
)

func file_pkg_echo_v1_echo_proto_rawDescGZIP() []byte {
	file_pkg_echo_v1_echo_proto_rawDescOnce.Do(func() {
		file_pkg_echo_v1_echo_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_echo_v1_echo_proto_rawDescData)
	})
	return file_pkg_echo_v1_echo_proto_rawDescData
}

var file_pkg_echo_v1_echo_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_pkg_echo_v1_echo_proto_goTypes = []interface{}{
	(*EchoRequest)(nil),    // 0: pkg.echo.v1.EchoRequest
	(*EchoResponse)(nil),   // 1: pkg.echo.v1.EchoResponse
	(*KubernetesInfo)(nil), // 2: pkg.echo.v1.KubernetesInfo
	(*RequestInfo)(nil),    // 3: pkg.echo.v1.RequestInfo
	(*ParsedURL)(nil),      // 4: pkg.echo.v1.ParsedURL
	(*KeyMultivalue)(nil),  // 5: pkg.echo.v1.KeyMultivalue
	(*RuntimeInfo)(nil),    // 6: pkg.echo.v1.RuntimeInfo
	(*StatusRequest)(nil),  // 7: pkg.echo.v1.StatusRequest
	(*StatusResponse)(nil), // 8: pkg.echo.v1.StatusResponse
}
var file_pkg_echo_v1_echo_proto_depIdxs = []int32{
	2, // 0: pkg.echo.v1.EchoResponse.kubernetes:type_name -> pkg.echo.v1.KubernetesInfo
	3, // 1: pkg.echo.v1.EchoResponse.request:type_name -> pkg.echo.v1.RequestInfo
	6, // 2: pkg.echo.v1.EchoResponse.runtime:type_name -> pkg.echo.v1.RuntimeInfo
	4, // 3: pkg.echo.v1.RequestInfo.parsed_url:type_name -> pkg.echo.v1.ParsedURL
	5, // 4: pkg.echo.v1.RequestInfo.header:type_name -> pkg.echo.v1.KeyMultivalue
	5, // 5: pkg.echo.v1.ParsedURL.query:type_name -> pkg.echo.v1.KeyMultivalue
	0, // 6: pkg.echo.v1.Echoer.Echo:input_type -> pkg.echo.v1.EchoRequest
	7, // 7: pkg.echo.v1.Echoer.Status:input_type -> pkg.echo.v1.StatusRequest
	1, // 8: pkg.echo.v1.Echoer.Echo:output_type -> pkg.echo.v1.EchoResponse
	8, // 9: pkg.echo.v1.Echoer.Status:output_type -> pkg.echo.v1.StatusResponse
	8, // [8:10] is the sub-list for method output_type
	6, // [6:8] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_pkg_echo_v1_echo_proto_init() }
func file_pkg_echo_v1_echo_proto_init() {
	if File_pkg_echo_v1_echo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_echo_v1_echo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EchoRequest); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EchoResponse); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KubernetesInfo); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestInfo); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParsedURL); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeyMultivalue); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RuntimeInfo); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusRequest); i {
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
		file_pkg_echo_v1_echo_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatusResponse); i {
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
			RawDescriptor: file_pkg_echo_v1_echo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_echo_v1_echo_proto_goTypes,
		DependencyIndexes: file_pkg_echo_v1_echo_proto_depIdxs,
		MessageInfos:      file_pkg_echo_v1_echo_proto_msgTypes,
	}.Build()
	File_pkg_echo_v1_echo_proto = out.File
	file_pkg_echo_v1_echo_proto_rawDesc = nil
	file_pkg_echo_v1_echo_proto_goTypes = nil
	file_pkg_echo_v1_echo_proto_depIdxs = nil
}
