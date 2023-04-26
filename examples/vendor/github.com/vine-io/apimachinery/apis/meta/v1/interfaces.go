// MIT License
//
// Copyright (c) 2021 Lack
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
	"github.com/vine-io/apimachinery/schema"
)

var _ schema.ObjectKind = (*TypeMeta)(nil)

func (m *TypeMeta) GetObjectKind() schema.ObjectKind {
	return m
}

func (m *TypeMeta) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	m.ApiVersion = gvk.APIGroup()
	m.Kind = gvk.Kind
}

func (m *TypeMeta) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromGVK(m.ApiVersion + "." + m.Kind)
}

type Meta interface {
	GetName() string
	SetName(name string)
	GetUID() any    // int or string
	SetUID(uid any) // int or string
	GetResourceVersion() string
	SetResourceVersion(rv string)
	GetNamespace() string
	SetNamespace(ns string)
	GetCreationTimestamp() int64
	SetCreationTimestamp(t int64)
	GetUpdateTimestamp() int64
	SetUpdateTimestamp(t int64)
	GetDeletionTimestamp() int64
	SetDeletionTimestamp(t int64)
	GetLabels() map[string]string
	SetLabels(labels map[string]string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
	GetGenerateName() string
	SetGenerateName(cn string)
	GetReferences() []*OwnerReference
	SetReferences(references []*OwnerReference)
}

var _ Meta = (*ObjectMeta)(nil)

func (m *ObjectMeta) GetName() string {
	return m.Name
}

func (m *ObjectMeta) SetName(name string) {
	m.Name = name
}

func (m *ObjectMeta) GetUID() any {
	return m.Uid
}

func (m *ObjectMeta) SetUID(uid any) {
	m.Uid = uid.(string)
}

func (m *ObjectMeta) GetResourceVersion() string {
	return m.ResourceVersion
}

func (m *ObjectMeta) SetResourceVersion(rv string) {
	m.ResourceVersion = rv
}

func (m *ObjectMeta) GetNamespace() string {
	return m.Namespace
}

func (m *ObjectMeta) SetNamespace(ns string) {
	m.Namespace = ns
}

func (m *ObjectMeta) GetCreationTimestamp() int64 {
	return m.CreationTimestamp
}

func (m *ObjectMeta) SetCreationTimestamp(t int64) {
	m.CreationTimestamp = t
}

func (m *ObjectMeta) GetUpdateTimestamp() int64 {
	return m.UpdateTimestamp
}

func (m *ObjectMeta) SetUpdateTimestamp(t int64) {
	m.UpdateTimestamp = t
}

func (m *ObjectMeta) GetDeletionTimestamp() int64 {
	return m.DeletionTimestamp
}

func (m *ObjectMeta) SetDeletionTimestamp(t int64) {
	m.DeletionTimestamp = t
}

func (m *ObjectMeta) GetLabels() map[string]string {
	return m.Labels
}

func (m *ObjectMeta) SetLabels(labels map[string]string) {
	m.Labels = labels
}

func (m *ObjectMeta) GetAnnotations() map[string]string {
	return m.Annotations
}

func (m *ObjectMeta) SetAnnotations(annotations map[string]string) {
	m.Annotations = annotations
}

func (m *ObjectMeta) GetGenerateName() string {
	return m.GenerateName
}

func (m *ObjectMeta) SetGenerateName(cn string) {
	m.GenerateName = cn
}

func (m *ObjectMeta) GetReferences() []*OwnerReference {
	return m.References
}

func (m *ObjectMeta) SetReferences(references []*OwnerReference) {
	m.References = references
}

func (m *ObjectMeta) PrimaryKey() (string, any, bool) {
	return "uid", m.Uid, m.Uid == ""
}

var _ Meta = (*EntityMeta)(nil)

func (m *EntityMeta) GetName() string {
	return m.Name
}

func (m *EntityMeta) SetName(name string) {
	m.Name = name
}

func (m *EntityMeta) GetUID() any {
	return m.Uid
}

func (m *EntityMeta) SetUID(uid any) {
	m.Uid = uid.(int64)
}

func (m *EntityMeta) GetResourceVersion() string {
	return m.ResourceVersion
}

func (m *EntityMeta) SetResourceVersion(rv string) {
	m.ResourceVersion = rv
}

func (m *EntityMeta) GetNamespace() string {
	return m.Namespace
}

func (m *EntityMeta) SetNamespace(ns string) {
	m.Namespace = ns
}

func (m *EntityMeta) GetCreationTimestamp() int64 {
	return m.CreationTimestamp
}

func (m *EntityMeta) SetCreationTimestamp(t int64) {
	m.CreationTimestamp = t
}

func (m *EntityMeta) GetUpdateTimestamp() int64 {
	return m.UpdateTimestamp
}

func (m *EntityMeta) SetUpdateTimestamp(t int64) {
	m.UpdateTimestamp = t
}

func (m *EntityMeta) GetDeletionTimestamp() int64 {
	return m.DeletionTimestamp
}

func (m *EntityMeta) SetDeletionTimestamp(t int64) {
	m.DeletionTimestamp = t
}

func (m *EntityMeta) GetLabels() map[string]string {
	return m.Labels
}

func (m *EntityMeta) SetLabels(labels map[string]string) {
	m.Labels = labels
}

func (m *EntityMeta) GetAnnotations() map[string]string {
	return m.Annotations
}

func (m *EntityMeta) SetAnnotations(annotations map[string]string) {
	m.Annotations = annotations
}

func (m *EntityMeta) GetGenerateName() string {
	return m.GenerateName
}

func (m *EntityMeta) SetGenerateName(cn string) {
	m.GenerateName = cn
}

func (m *EntityMeta) GetReferences() []*OwnerReference {
	return m.References
}

func (m *EntityMeta) SetReferences(references []*OwnerReference) {
	m.References = references
}

func (m *EntityMeta) PrimaryKey() (string, any, bool) {
	return "uid", m.Uid, m.Uid == 0
}

/*
type ListMeta struct {
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,1,opt,name=resourceVersion"`
	Page            int32  `json:"page,omitempty" protobuf:"varint,2,opt,name=page"`
	Size            int32  `json:"size,omitempty" protobuf:"varint,3,opt,name=size"`
	Total           int64  `json:"total,omitempty" protobuf:"varint,4,opt,name=total"`
}
*/

var _ Lister = (*ListMeta)(nil)

type Lister interface {
	GetResourceVersion() string
	SetResourceVersion(version string)
	GetPage() int32
	SetPage(page int32)
	GetSize() int32
	SetSize(s int32)
	GetTotal() int64
	SetTotal(total int64)
}

func (m *ListMeta) GetResourceVersion() string {
	return m.ResourceVersion
}

func (m *ListMeta) SetResourceVersion(version string) {
	m.ResourceVersion = version
}

func (m *ListMeta) GetPage() int32 {
	return m.Page
}

func (m *ListMeta) SetPage(page int32) {
	m.Page = page
}

func (m *ListMeta) GetSize() int32 {
	return m.Size
}

func (m *ListMeta) SetSize(s int32) {
	m.Size = s
}

func (m *ListMeta) GetTotal() int64 {
	return m.Total
}

func (m *ListMeta) SetTotal(total int64) {
	m.Total = total
}
