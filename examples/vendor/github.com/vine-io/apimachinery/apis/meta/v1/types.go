// MIT License
//
// Copyright (c) 2023 Lack
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package v1

import (
	"database/sql/driver"

	"github.com/vine-io/apimachinery/storage/dao"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type StatusCode int32

const (
	StatusNoReady StatusCode = 0
	StatusReady   StatusCode = 1
)

// +gogo:deepcopy=true
// +gogo:genproto=true
// +gogo:gengorm=true
type TypeMeta struct {
	// 资源类型
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind,proto3"`
	// 资源版本信息
	ApiVersion string `json:"apiVersion,omitempty" protobuf:"bytes,2,opt,name=apiVersion,proto3"`
}

// +gogo:deepcopy=true
// +gogo:genproto=true
type OwnerReference struct {
	ApiVersion string `json:"apiVersion,omitempty" protobuf:"bytes,1,opt,name=apiVersion,proto3"`
	Kind       string `json:"kind,omitempty" protobuf:"bytes,2,opt,name=kind,proto3"`
	Name       string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name,proto3"`
	Uid        string `json:"uid,omitempty" protobuf:"bytes,4,opt,name=uid,proto3"`
}

// Value return json value, implement driver.Valuer interface
func (m *OwnerReference) Value() (driver.Value, error) {
	return dao.GetValue(m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *OwnerReference) Scan(value any) error {
	return dao.ScanValue(value, m)
}

// GormDBDataType implements migrator.GormDBDataTypeInterface interface
func (m *OwnerReference) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return dao.GetGormDBDataType(db, field)
}

// +gogo:deepcopy=true
// +gogo:genproto=true
// +gogo:gengorm=true
type EntityMeta struct {
	// 资源名称
	Name string `json:"name" protobuf:"bytes,1,opt,name=name,proto3"`
	// 资源 id
	// +primaryKey
	// +gen:primaryKey
	Uid int64 `json:"uid" protobuf:"varint,2,opt,name=uid,proto3"`
	// 资源版本信息
	ResourceVersion string `json:"resourceVersion" protobuf:"bytes,3,opt,name=resourceVersion,proto3"`
	// 资源描述信息
	Description string `json:"description" protobuf:"bytes,4,opt,name=description,proto3"`
	// 资源命名空间
	Namespace string `json:"namespace" protobuf:"bytes,5,opt,name=namespace,proto3"`
	// 资源创建时间
	CreationTimestamp int64 `json:"creationTimestamp" protobuf:"varint,6,opt,name=creationTimestamp,proto3"`
	// 资源更新时间
	UpdateTimestamp int64 `json:"updateTimestamp" protobuf:"varint,7,opt,name=updateTimestamp,proto3"`
	// 删除时间
	DeletionTimestamp int64  `json:"deletionTimestamp" protobuf:"varint,8,opt,name=deletionTimestamp,proto3"`
	GenerateName      string `json:"generateName" protobuf:"bytes,9,opt,name=generateName,proto3"`
	// 资源标签
	Labels dao.Map[string, string] `json:"labels" protobuf:"bytes,10,rep,name=labels,proto3,casttype=github.com/vine-io/apimachinery/storage/dao.Map"`
	// 资源注解
	// 只读
	Annotations dao.Map[string, string] `json:"annotations" protobuf:"bytes,11,rep,name=annotations,proto3,casttype=github.com/vine-io/apimachinery/storage/dao.Map"`
	// 资源关系信息
	// 该资源示例能稳定运行的依赖资源
	References dao.JSONArray[*OwnerReference] `json:"references" protobuf:"bytes,12,rep,name=references,proto3"`
}

// +gogo:deepcopy=true
// +gogo:genproto=true
// +gogo:gengorm=true
type ObjectMeta struct {
	// 资源名称
	Name string `json:"name" protobuf:"bytes,1,opt,name=name,proto3"`
	// 资源 id
	// +primaryKey
	// +gen:primaryKey
	Uid string `json:"uid" protobuf:"bytes,2,opt,name=uid,proto3"`
	// 资源版本信息
	ResourceVersion string `json:"resourceVersion" protobuf:"bytes,3,opt,name=resourceVersion,proto3"`
	// 资源描述信息
	Description string `json:"description" protobuf:"bytes,4,opt,name=description,proto3"`
	// 资源命名空间
	Namespace string `json:"namespace" protobuf:"bytes,5,opt,name=namespace,proto3"`
	// 资源创建时间
	CreationTimestamp int64 `json:"creationTimestamp" protobuf:"varint,6,opt,name=creationTimestamp,proto3"`
	// 资源更新时间
	UpdateTimestamp int64 `json:"updateTimestamp" protobuf:"varint,7,opt,name=updateTimestamp,proto3"`
	// 删除时间
	DeletionTimestamp int64  `json:"deletionTimestamp" protobuf:"varint,8,opt,name=deletionTimestamp,proto3"`
	GenerateName      string `json:"generateName" protobuf:"bytes,9,opt,name=generateName,proto3"`
	// 资源标签
	Labels dao.Map[string, string] `json:"labels" protobuf:"bytes,10,rep,name=labels,proto3,casttype=github.com/vine-io/apimachinery/storage/dao.Map"`
	// 资源注解
	// 只读
	Annotations dao.Map[string, string] `json:"annotations" protobuf:"bytes,11,rep,name=annotations,proto3,casttype=github.com/vine-io/apimachinery/storage/dao.Map"`
	// 资源关系信息
	// 该资源示例能稳定运行的依赖资源
	References dao.JSONArray[*OwnerReference] `json:"references" protobuf:"bytes,12,rep,name=references,proto3"`
}

// Value return json value, implement driver.Valuer interface
func (m *ObjectMeta) Value() (driver.Value, error) {
	return dao.GetValue(m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *ObjectMeta) Scan(value any) error {
	return dao.ScanValue(value, m)
}

// GormDBDataType implements migrator.GormDBDataTypeInterface interface
func (m *ObjectMeta) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return dao.GetGormDBDataType(db, field)
}

// +gogo:deepcopy=true
// +gogo:genproto=true
type ListMeta struct {
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,1,opt,name=resourceVersion,proto3"`
	Page            int32  `json:"page,omitempty" protobuf:"varint,2,opt,name=page,proto3"`
	Size            int32  `json:"size,omitempty" protobuf:"varint,3,opt,name=size,proto3"`
	Total           int64  `json:"total,omitempty" protobuf:"varint,4,opt,name=total,proto3"`
}

// +gogo:deepcopy=true
// +gogo:genproto=true
// 资源状态
type State struct {
	Code StatusCode `json:"code,omitempty" protobuf:"varint,1,opt,name=code,proto3,casttype=StatusCode"`
	// code != 0 时，显示错误信息
	Message string `json:"message,omitempty" protobuf:"bytes,2,opt,name=message,proto3"`
}
