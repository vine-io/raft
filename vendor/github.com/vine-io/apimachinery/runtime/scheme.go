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
	"reflect"

	"github.com/vine-io/apimachinery/schema"
)

var DefaultScheme Scheme = NewScheme()

type SimpleScheme struct {
	gvkToTypes map[schema.GroupVersionKind]reflect.Type

	typesToGvk map[reflect.Type]schema.GroupVersionKind

	defaultFuncs map[reflect.Type]DefaultFunc

	gFn DefaultFunc

	observedVersions []schema.GroupVersion
}

// New creates a new Object, and call global DefaultFunc
func (s *SimpleScheme) New(gvk schema.GroupVersionKind) (Object, error) {
	rv, exists := s.gvkToTypes[gvk]
	if !exists {
		return nil, ErrUnknownGVK
	}

	out := reflect.New(rv).Interface().(Object)
	out.GetObjectKind().SetGroupVersionKind(gvk)

	if s.gFn != nil {
		out = s.gFn(out, gvk)
	}

	return out, nil
}

// IsExists checks schema.GroupVersionKind exists
func (s *SimpleScheme) IsExists(gvk schema.GroupVersionKind) bool {
	_, ok := s.gvkToTypes[gvk]
	return ok
}

// AllGVKs returns all schema.GroupVersionKind
func (s *SimpleScheme) AllGVKs() []schema.GroupVersionKind {
	gvks := make([]schema.GroupVersionKind, 0)
	for gvk, _ := range s.gvkToTypes {
		gvks = append(gvks, gvk)
	}
	return gvks
}

// AllObjects returns all Object
func (s *SimpleScheme) AllObjects() []Object {
	objects := make([]Object, 0)
	for gvk, rv := range s.gvkToTypes {
		out := reflect.New(rv).Interface().(Object)
		out.GetObjectKind().SetGroupVersionKind(gvk)
		objects = append(objects, out)
	}
	return objects
}

// AddKnownTypes add Object to Scheme
func (s *SimpleScheme) AddKnownTypes(gv schema.GroupVersion, types ...Object) error {
	s.addObservedVersion(gv)
	for _, v := range types {
		rt := reflect.TypeOf(v)
		if rt.Kind() != reflect.Ptr {
			return ErrIsNotPointer
		}
		rt = rt.Elem()
		gvk := gv.WithKind(rt.Name())
		s.gvkToTypes[gvk] = rt
		s.typesToGvk[rt] = gvk
	}

	return nil
}

// Default call global DefaultFunc specifies DefaultFunc
func (s *SimpleScheme) Default(src Object) Object {
	rt := reflect.TypeOf(src)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	gvk := s.typesToGvk[rt]
	if src.GetObjectKind().GroupVersionKind().Empty() {
		src.GetObjectKind().SetGroupVersionKind(gvk)
	}

	if s.gFn != nil {
		src = s.gFn(src, gvk)
	}

	fn, exists := s.defaultFuncs[rt]
	if exists {
		src = fn(src, gvk)
	}
	return src
}

func (s *SimpleScheme) AddTypeDefaultingFunc(srcType Object, fn DefaultFunc) {
	s.defaultFuncs[reflect.TypeOf(srcType)] = fn
}

func (s *SimpleScheme) AddGlobalDefaultingFunc(fn DefaultFunc) {
	s.gFn = fn
}

func (s *SimpleScheme) addObservedVersion(gv schema.GroupVersion) {
	if gv.Version == "" {
		return
	}

	for _, observedVersion := range s.observedVersions {
		if observedVersion == gv {
			return
		}
	}

	s.observedVersions = append(s.observedVersions, gv)
}

func NewScheme() *SimpleScheme {
	return &SimpleScheme{
		gvkToTypes:       map[schema.GroupVersionKind]reflect.Type{},
		typesToGvk:       map[reflect.Type]schema.GroupVersionKind{},
		defaultFuncs:     map[reflect.Type]DefaultFunc{},
		observedVersions: []schema.GroupVersion{},
	}
}

// NewObject calls DefaultScheme.New()
func NewObject(gvk schema.GroupVersionKind) (Object, error) {
	return DefaultScheme.New(gvk)
}

// IsExists calls DefaultScheme.IsExists()
func IsExists(gvk schema.GroupVersionKind) bool {
	return DefaultScheme.IsExists(gvk)
}

// AllGVKs calls DefaultScheme.AllGVKs()
func AllGVKs() []schema.GroupVersionKind {
	return DefaultScheme.AllGVKs()
}

// AllObjects calls DefaultScheme.AllObjects()
func AllObjects() []Object {
	return DefaultScheme.AllObjects()
}

// AddKnownTypes calls DefaultScheme.AddKnownTypes()
func AddKnownTypes(gv schema.GroupVersion, types ...Object) error {
	return DefaultScheme.AddKnownTypes(gv, types...)
}

// DefaultObject calls DefaultScheme.Default()
func DefaultObject(src Object) Object {
	return DefaultScheme.Default(src)
}

// AddGlobalDefaultingFunc calls DefaultScheme.AddTypeDefaultingFunc()
func AddGlobalDefaultingFunc(fn DefaultFunc) {
	DefaultScheme.AddGlobalDefaultingFunc(fn)
}

// AddTypeDefaultingFunc calls DefaultScheme.AddTypeDefaultingFunc()
func AddTypeDefaultingFunc(srcType Object, fn DefaultFunc) {
	DefaultScheme.AddTypeDefaultingFunc(srcType, fn)
}
