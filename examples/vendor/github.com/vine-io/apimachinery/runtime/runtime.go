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

package runtime

import (
	"github.com/vine-io/apimachinery/schema"
)

// Object interface must be supported by all API types registered with Scheme. Since objects in a scheme are
// expected to be serialized to the wire, the interface an Object must provide to the Scheme allows
// serializers to set the kind, version, and group the object is represented as. An Object may choose
// to return a no-op ObjectKindAccessor in cases where it is not expected to be serialized.
type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopy() Object
	DeepFrom(Object)
}

type DefaultFunc func(src Object, gvk schema.GroupVersionKind) Object

type Scheme interface {
	// New creates a new object
	New(gvk schema.GroupVersionKind) (Object, error)

	// AllGVKs returns all schema.GroupVersionKind
	AllGVKs() []schema.GroupVersionKind

	// AllObjects returns all Objects
	AllObjects() []Object

	// IsExists checks whether the object exists
	IsExists(gvk schema.GroupVersionKind) bool

	// AddKnownTypes adds Objects to Machinery
	AddKnownTypes(gv schema.GroupVersion, types ...Object) error

	// Default calls the DefaultFunc to src
	Default(src Object) Object

	// AddGlobalDefaultingFunc adds global DefaultFunc to Machinery
	AddGlobalDefaultingFunc(fn DefaultFunc)

	// AddTypeDefaultingFunc adds DefaultFunc to Machinery
	AddTypeDefaultingFunc(srcType Object, fn DefaultFunc)
}
