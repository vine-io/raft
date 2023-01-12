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

package metav1

import "github.com/vine-io/apimachinery/schema"

var _ Meta = (*ObjectMeta)(nil)

type Meta interface {
	GetName() string
	SetName(name string)
	GetUID() string
	SetUID(uid string)
	GetResourceVersion() int64
	SetResourceVersion(rv int64)
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
	GetClusterName() string
	SetClusterName(cn string)
	GetReferences() []*OwnerReference
	SetReferences(references []*OwnerReference)
}

func (m *ObjectMeta) GetName() string {
	return m.Name
}

func (m *ObjectMeta) SetName(name string) {
	m.Name = name
}

func (m *ObjectMeta) GetUID() string {
	return m.Uid
}

func (m *ObjectMeta) SetUID(uid string) {
	m.Uid = uid
}

func (m *ObjectMeta) GetResourceVersion() int64 {
	return m.ResourceVersion
}

func (m *ObjectMeta) SetResourceVersion(rv int64) {
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

func (m *ObjectMeta) GetClusterName() string {
	return m.ClusterName
}

func (m *ObjectMeta) SetClusterName(cn string) {
	m.ClusterName = cn
}

func (m *ObjectMeta) GetReferences() []*OwnerReference {
	return m.References
}

func (m *ObjectMeta) SetReferences(references []*OwnerReference) {
	m.References = references
}

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
