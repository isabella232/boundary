// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.12.4
// source: controller/storage/authtoken/store/v1/authtoken.proto

package store

import (
	timestamp "github.com/hashicorp/boundary/internal/db/timestamp"
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

type AuthToken struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// public_id is used to access the auth token via an API
	// @inject_tag: gorm:"primary_key"
	PublicId string `protobuf:"bytes,1,opt,name=public_id,json=publicId,proto3" json:"public_id,omitempty" gorm:"primary_key"`
	// create_time from the RDBMS
	// @inject_tag: `gorm:"default:current_timestamp"`
	CreateTime *timestamp.Timestamp `protobuf:"bytes,2,opt,name=create_time,json=createTime,proto3" json:"create_time,omitempty" gorm:"default:current_timestamp"`
	// update_time from the RDBMS
	// @inject_tag: `gorm:"default:current_timestamp"`
	UpdateTime *timestamp.Timestamp `protobuf:"bytes,3,opt,name=update_time,json=updateTime,proto3" json:"update_time,omitempty" gorm:"default:current_timestamp"`
	// last_access_time indicates the last time the auth token was used on the boundary API.
	// @inject_tag: `gorm:"default:current_timestamp"`
	ApproximateLastAccessTime *timestamp.Timestamp `protobuf:"bytes,4,opt,name=approximate_last_access_time,json=approximateLastAccessTime,proto3" json:"approximate_last_access_time,omitempty" gorm:"default:current_timestamp"`
	// expiration_time indicates when this session will expire.
	// If null a default duration and create_time is used to calculate expiration.
	// @inject_tag: `gorm:"default:null"`
	ExpirationTime *timestamp.Timestamp `protobuf:"bytes,5,opt,name=expiration_time,json=expirationTime,proto3" json:"expiration_time,omitempty" gorm:"default:null"`
	// ciphertext token value stored in the database
	// @inject_tag: gorm:"column:token;not_null" wrapping:"ct,authtoken_token"
	CtToken []byte `protobuf:"bytes,6,opt,name=ct_token,json=ctToken,proto3" json:"ct_token,omitempty" gorm:"column:token;not_null" wrapping:"ct,authtoken_token"`
	// plain text version of the decrypted authtoken value
	// we are NOT storing this plain-text entry data in the db
	// token is the field stored and used by the client
	// @inject_tag: gorm:"-" wrapping:"pt,authtoken_token"
	Token string `protobuf:"bytes,7,opt,name=token,proto3" json:"token,omitempty" gorm:"-" wrapping:"pt,authtoken_token"`
	// auth_account_id is the public id for the auth account this auth token
	// was generated for.
	// @inject_tag: `gorm:"default:not_null"`
	AuthAccountId string `protobuf:"bytes,10,opt,name=auth_account_id,json=authAccountId,proto3" json:"auth_account_id,omitempty" gorm:"default:not_null"`
	// scope_id is not stored in the backing DB but it derived from the linked to auth account.
	// @inject_tag: gorm:"-"
	ScopeId string `protobuf:"bytes,11,opt,name=scope_id,json=scopeId,proto3" json:"scope_id,omitempty" gorm:"-"`
	// auth_method_id is not stored in the backing DB but it derived from the linked to auth account.
	// @inject_tag: gorm:"-"
	AuthMethodId string `protobuf:"bytes,12,opt,name=auth_method_id,json=authMethodId,proto3" json:"auth_method_id,omitempty" gorm:"-"`
	// iam_user_id is not stored in the backing DB but it derived from the linked to auth account.
	// @inject_tag: gorm:"-"
	IamUserId string `protobuf:"bytes,13,opt,name=iam_user_id,json=iamUserId,proto3" json:"iam_user_id,omitempty" gorm:"-"`
	// key_id is the key ID that was used for the encryption operation. It can be
	// used to identify a specific version of the key needed to decrypt the value,
	// which is useful for caching purposes.
	// @inject_tag: `gorm:"not_null"`
	KeyId string `protobuf:"bytes,14,opt,name=key_id,json=keyId,proto3" json:"key_id,omitempty" gorm:"not_null"`
}

func (x *AuthToken) Reset() {
	*x = AuthToken{}
	if protoimpl.UnsafeEnabled {
		mi := &file_controller_storage_authtoken_store_v1_authtoken_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthToken) ProtoMessage() {}

func (x *AuthToken) ProtoReflect() protoreflect.Message {
	mi := &file_controller_storage_authtoken_store_v1_authtoken_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthToken.ProtoReflect.Descriptor instead.
func (*AuthToken) Descriptor() ([]byte, []int) {
	return file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescGZIP(), []int{0}
}

func (x *AuthToken) GetPublicId() string {
	if x != nil {
		return x.PublicId
	}
	return ""
}

func (x *AuthToken) GetCreateTime() *timestamp.Timestamp {
	if x != nil {
		return x.CreateTime
	}
	return nil
}

func (x *AuthToken) GetUpdateTime() *timestamp.Timestamp {
	if x != nil {
		return x.UpdateTime
	}
	return nil
}

func (x *AuthToken) GetApproximateLastAccessTime() *timestamp.Timestamp {
	if x != nil {
		return x.ApproximateLastAccessTime
	}
	return nil
}

func (x *AuthToken) GetExpirationTime() *timestamp.Timestamp {
	if x != nil {
		return x.ExpirationTime
	}
	return nil
}

func (x *AuthToken) GetCtToken() []byte {
	if x != nil {
		return x.CtToken
	}
	return nil
}

func (x *AuthToken) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *AuthToken) GetAuthAccountId() string {
	if x != nil {
		return x.AuthAccountId
	}
	return ""
}

func (x *AuthToken) GetScopeId() string {
	if x != nil {
		return x.ScopeId
	}
	return ""
}

func (x *AuthToken) GetAuthMethodId() string {
	if x != nil {
		return x.AuthMethodId
	}
	return ""
}

func (x *AuthToken) GetIamUserId() string {
	if x != nil {
		return x.IamUserId
	}
	return ""
}

func (x *AuthToken) GetKeyId() string {
	if x != nil {
		return x.KeyId
	}
	return ""
}

var File_controller_storage_authtoken_store_v1_authtoken_proto protoreflect.FileDescriptor

var file_controller_storage_authtoken_store_v1_authtoken_proto_rawDesc = []byte{
	0x0a, 0x35, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x73, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2f, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x74, 0x6f, 0x6b, 0x65,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x25, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c,
	0x6c, 0x65, 0x72, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x61, 0x75, 0x74, 0x68,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x2f,
	0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2f, 0x76, 0x31, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xd5, 0x04, 0x0a, 0x09, 0x41, 0x75, 0x74, 0x68, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1b, 0x0a,
	0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x49, 0x64, 0x12, 0x4b, 0x0a, 0x0b, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x2a, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x73, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x2e, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x4b, 0x0a, 0x0b, 0x75, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x2e, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x54, 0x69, 0x6d, 0x65, 0x12, 0x6b, 0x0a, 0x1c, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x78, 0x69, 0x6d,
	0x61, 0x74, 0x65, 0x5f, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x19, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x78, 0x69, 0x6d,
	0x61, 0x74, 0x65, 0x4c, 0x61, 0x73, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x53, 0x0a, 0x0f, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0e, 0x65, 0x78, 0x70, 0x69, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x74, 0x5f, 0x74, 0x6f, 0x6b,
	0x65, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63, 0x74, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x26, 0x0a, 0x0f, 0x61, 0x75, 0x74, 0x68, 0x5f,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x61, 0x75, 0x74, 0x68, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x49, 0x64, 0x12,
	0x19, 0x0a, 0x08, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x49, 0x64, 0x12, 0x24, 0x0a, 0x0e, 0x61, 0x75,
	0x74, 0x68, 0x5f, 0x6d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x0c, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x61, 0x75, 0x74, 0x68, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x49, 0x64,
	0x12, 0x1e, 0x0a, 0x0b, 0x69, 0x61, 0x6d, 0x5f, 0x75, 0x73, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18,
	0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x69, 0x61, 0x6d, 0x55, 0x73, 0x65, 0x72, 0x49, 0x64,
	0x12, 0x15, 0x0a, 0x06, 0x6b, 0x65, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x6b, 0x65, 0x79, 0x49, 0x64, 0x42, 0x3e, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x61, 0x73, 0x68, 0x69, 0x63, 0x6f, 0x72, 0x70, 0x2f,
	0x62, 0x6f, 0x75, 0x6e, 0x64, 0x61, 0x72, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x2f, 0x73, 0x74, 0x6f, 0x72,
	0x65, 0x3b, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescOnce sync.Once
	file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescData = file_controller_storage_authtoken_store_v1_authtoken_proto_rawDesc
)

func file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescGZIP() []byte {
	file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescOnce.Do(func() {
		file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescData = protoimpl.X.CompressGZIP(file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescData)
	})
	return file_controller_storage_authtoken_store_v1_authtoken_proto_rawDescData
}

var file_controller_storage_authtoken_store_v1_authtoken_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_controller_storage_authtoken_store_v1_authtoken_proto_goTypes = []interface{}{
	(*AuthToken)(nil),           // 0: controller.storage.authtoken.store.v1.AuthToken
	(*timestamp.Timestamp)(nil), // 1: controller.storage.timestamp.v1.Timestamp
}
var file_controller_storage_authtoken_store_v1_authtoken_proto_depIdxs = []int32{
	1, // 0: controller.storage.authtoken.store.v1.AuthToken.create_time:type_name -> controller.storage.timestamp.v1.Timestamp
	1, // 1: controller.storage.authtoken.store.v1.AuthToken.update_time:type_name -> controller.storage.timestamp.v1.Timestamp
	1, // 2: controller.storage.authtoken.store.v1.AuthToken.approximate_last_access_time:type_name -> controller.storage.timestamp.v1.Timestamp
	1, // 3: controller.storage.authtoken.store.v1.AuthToken.expiration_time:type_name -> controller.storage.timestamp.v1.Timestamp
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_controller_storage_authtoken_store_v1_authtoken_proto_init() }
func file_controller_storage_authtoken_store_v1_authtoken_proto_init() {
	if File_controller_storage_authtoken_store_v1_authtoken_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_controller_storage_authtoken_store_v1_authtoken_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthToken); i {
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
			RawDescriptor: file_controller_storage_authtoken_store_v1_authtoken_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_controller_storage_authtoken_store_v1_authtoken_proto_goTypes,
		DependencyIndexes: file_controller_storage_authtoken_store_v1_authtoken_proto_depIdxs,
		MessageInfos:      file_controller_storage_authtoken_store_v1_authtoken_proto_msgTypes,
	}.Build()
	File_controller_storage_authtoken_store_v1_authtoken_proto = out.File
	file_controller_storage_authtoken_store_v1_authtoken_proto_rawDesc = nil
	file_controller_storage_authtoken_store_v1_authtoken_proto_goTypes = nil
	file_controller_storage_authtoken_store_v1_authtoken_proto_depIdxs = nil
}
