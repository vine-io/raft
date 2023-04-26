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

package storage

import (
	"fmt"
	"reflect"

	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/schema"
	"gorm.io/gorm"
)

var (
	ErrInvalidObject       = fmt.Errorf("specified object invalid")
	ErrStorageIsNotPointer = fmt.Errorf("storage is not a pointer")
	ErrStorageNotExists    = fmt.Errorf("storage not exists")
	ErrStorageAutoMigrate  = fmt.Errorf("auto migrate storage")
)

type GenericStorageFactory struct {
	gvkToType map[schema.GroupVersionKind]reflect.Type
}

func (s *GenericStorageFactory) AddKnownStorages(tx *gorm.DB, gv schema.GroupVersion, sets ...Storage) error {

	for _, storage := range sets {
		rt := reflect.TypeOf(storage)
		if rt.Kind() != reflect.Ptr {
			return ErrStorageIsNotPointer
		}
		rt = rt.Elem()
		gvk := gv.WithKind(storage.Target().Elem().Name())
		s.gvkToType[gvk] = rt

		if err := storage.AutoMigrate(tx); err != nil {
			return fmt.Errorf("%w: %v", ErrStorageAutoMigrate, err)
		}
	}

	return nil
}

func (s *GenericStorageFactory) NewStorage(tx *gorm.DB, in runtime.Object) (Storage, error) {
	gvk := in.GetObjectKind().GroupVersionKind()
	rt, exists := s.gvkToType[gvk]
	if !exists {
		return nil, fmt.Errorf("%w: object's gvk is %s", ErrStorageNotExists, in.GetObjectKind().GroupVersionKind())
	}

	storage := reflect.New(rt).Interface().(Storage)

	err := storage.Load(tx, in)
	if err != nil {
		return nil, fmt.Errorf("load object: %v", err)
	}

	return storage, nil
}

func (s *GenericStorageFactory) IsExists(gvk schema.GroupVersionKind) bool {
	_, ok := s.gvkToType[gvk]
	return ok
}

func (s *GenericStorageFactory) AllStorages() []Storage {
	storages := make([]Storage, 0)

	for _, rt := range s.gvkToType {
		storage := reflect.New(rt).Interface().(Storage)
		storages = append(storages, storage)
	}

	return storages
}

func NewStorageFactory() Factory {
	return &GenericStorageFactory{
		gvkToType: map[schema.GroupVersionKind]reflect.Type{},
	}
}
