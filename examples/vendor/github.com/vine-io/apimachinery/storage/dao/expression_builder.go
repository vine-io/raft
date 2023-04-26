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

package dao

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm/clause"
)

type EOp int32

const (
	EqOp EOp = iota + 1
	NeqOp
	GtOp
	GteOp
	LtOp
	LteOp
	InOp
	LikeOp
)

func ParseOp(v interface{}) (op EOp) {
	switch v.(type) {
	case string:
		vv := v.(string)
		if strings.HasPrefix(vv, "%") || strings.HasSuffix(vv, "%") {
			op = LikeOp
		} else {
			op = EqOp
		}
	default:
		op = EqOp
	}
	return
}

// condExpr the builder for interface Expression
type condExpr struct {
	op EOp
}

func Cond() *condExpr {
	return &condExpr{op: EqOp}
}

func (e *condExpr) Op(op EOp) *condExpr {
	e.op = op
	return e
}

func (e *condExpr) Build(c string, v interface{}, vv ...interface{}) clause.Expression {
	switch e.op {
	case EqOp:
		return clause.Eq{Column: clause.Column{Name: c}, Value: v}
	case NeqOp:
		return clause.Neq{Column: clause.Column{Name: c}, Value: v}
	case GtOp:
		return clause.Gt{Column: clause.Column{Name: c}, Value: v}
	case GteOp:
		return clause.Gte{Column: clause.Column{Name: c}, Value: v}
	case LtOp:
		return clause.Lt{Column: clause.Column{Name: c}, Value: v}
	case LteOp:
		return clause.Lte{Column: clause.Column{Name: c}, Value: v}
	case InOp:
		return clause.IN{Column: clause.Column{Name: c}, Values: append([]interface{}{v}, vv...)}
	case LikeOp:
		return clause.Like{Column: clause.Column{Name: c}, Value: v}
	}
	return nil
}

func patch(obj reflect.Value, outs map[string]interface{}, key string) {
	vo := reflectDirect(obj)
	switch vo.Kind() {
	case reflect.String:
		vv := vo.String()
		if vv != "" {
			outs[key] = vv
		}
	case reflect.Float32, reflect.Float64:
		vv := vo.Float()
		if vv != 0 {
			outs[key] = vv
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vo.Int()
		if vv != 0 {
			outs[key] = vv
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vo.Uint()
		if vv != 0 {
			outs[key] = vv
		}
	case reflect.Map:
		for _, mkv := range vo.MapKeys() {
			var kk string
			switch mkv.Type().Kind() {
			case reflect.String:
				kk = key + "." + mkv.String()
			case reflect.Float32, reflect.Float64:
				kk = key + "." + fmt.Sprintf("%f", mkv.Float())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				kk = key + "." + fmt.Sprintf("%d", mkv.Int())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				kk = key + "." + fmt.Sprintf("%d", mkv.Uint())
			default:
				continue
			}
			mvv := vo.MapIndex(mkv)
			switch mvv.Type().Kind() {
			case reflect.String:
				outs[kk] = mvv.String()
			case reflect.Float32, reflect.Float64:
				outs[kk] = mvv.Float()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				outs[kk] = mvv.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				outs[kk] = mvv.Uint()
			case reflect.Map:
				patch(mvv, outs, kk)
			case reflect.Ptr:
				mvv = reflectDirect(mvv)
				fallthrough
			case reflect.Struct:
				if !mvv.IsValid() {
					continue
				}
				if mvv.IsZero() {
					outs[kk] = nil
					continue
				}

				for i := 0; i < mvv.NumField(); i++ {
					mvvd := mvv.Type().Field(i)
					mvfdv := mvv.Field(i)
					jsonName := strings.Split(mvvd.Tag.Get("json"), ",")[0]
					patch(mvfdv, outs, kk+"."+jsonName)
				}
			}

		}
	case reflect.Struct:
		if !vo.IsValid() {
			return
		}
		if vo.IsZero() {
			outs[key] = nil
			return
		}
		for i := 0; i < vo.Type().NumField(); i++ {
			kk := ""
			fd := vo.Type().Field(i)
			fdv := vo.Field(i)

			if !fdv.IsValid() {
				continue
			}

			if fdv.IsZero() {
				continue
			}

			jsonName := strings.Split(fd.Tag.Get("json"), ",")[0]
			if key == "" {
				kk = jsonName
			} else {
				kk = key + "." + jsonName
			}

			patch(fdv, outs, kk)
		}
	}
}

func reflectDirect(v reflect.Value) reflect.Value {
	if v.Type().Kind() == reflect.Ptr {
		for v.Type().Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}
	return v
}

func FieldPatch(v interface{}) map[string]interface{} {
	outs := make(map[string]interface{})
	patch(reflect.ValueOf(v), outs, "")
	return outs
}
