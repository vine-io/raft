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
	"context"
	"reflect"

	"github.com/vine-io/apimachinery/runtime"
	"github.com/vine-io/apimachinery/schema"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Storage interface {
	Target() reflect.Type
	AutoMigrate(tx *gorm.DB) error
	Load(tx *gorm.DB, object runtime.Object) error
	FindPage(ctx context.Context, page, size int32) (runtime.Object, error)
	FindAll(ctx context.Context) (runtime.Object, error)
	Count(ctx context.Context) (total int64, err error)
	FindPk(ctx context.Context, pk any) (runtime.Object, error)
	FindOne(ctx context.Context) (runtime.Object, error)
	Cond(exprs ...clause.Expression) Storage
	Create(ctx context.Context) (runtime.Object, error)
	Updates(ctx context.Context) (runtime.Object, error)
	Delete(ctx context.Context, soft bool) error
}

type Factory interface {
	// AddKnownStorages registers Storages
	AddKnownStorages(tx *gorm.DB, gv schema.GroupVersion, sets ...Storage) error

	// NewStorage get a Storage by runtime.Object
	NewStorage(tx *gorm.DB, in runtime.Object) (Storage, error)

	// IsExists checks Storage exists
	IsExists(gvk schema.GroupVersionKind) bool

	// AllStorages returns all Storages
	AllStorages() []Storage
}
