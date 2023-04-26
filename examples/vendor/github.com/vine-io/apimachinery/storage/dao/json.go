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
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	json "github.com/json-iterator/go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type JSONValue interface {
	sql.Scanner
	driver.Valuer
	GormDBDataType(db *gorm.DB, field *schema.Field) string
}

type Builtin interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~complex64 | ~complex128 |
		~float32 | ~float64 |
		~string | ~uintptr
}

type Array[V Builtin] []V

func (m *Array[V]) PushBack(value V) {
	*m = append(*m, value)
}

func (m *Array[V]) PopBack() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	v = (*m)[n-1]
	*m = (*m)[:n-1]
	return
}

func (m *Array[V]) PushFront(value V) {
	*m = append([]V{value}, *m...)
}

func (m *Array[V]) PopFront() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	v = (*m)[0]
	*m = (*m)[1:]
	return
}

func (m *Array[V]) Front() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	return (*m)[0], true
}

func (m *Array[V]) Back() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	return (*m)[n-1], true
}

func (m *Array[V]) Get(idx int) (v V, ok bool) {
	if idx < 0 || idx >= len(*m) {
		return v, false
	}

	return (*m)[idx], true
}

func (m *Array[V]) Remove(idx int) (v V, ok bool) {
	if idx < 0 || idx >= len(*m) {
		return v, false
	}

	if idx == 0 {
		return m.PopFront()
	}
	if idx == len(*m)-1 {
		return m.PopBack()
	}

	v = (*m)[idx]
	*m = append((*m)[0:idx], (*m)[idx+1:]...)

	return v, true
}

// Value return json value, implement driver.Valuer interface
func (m *Array[V]) Value() (driver.Value, error) {
	return GetValue(m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *Array[V]) Scan(value any) error {
	return ScanValue(value, m)
}

// GormDBDataType implements migrator.GormDBDataTypeInterface interface
func (m *Array[V]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return GetGormDBDataType(db, field)
}

type JSONArray[V JSONValue] []V

func (m *JSONArray[V]) PushBack(value V) {
	*m = append(*m, value)
}

func (m *JSONArray[V]) PopBack() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	v = (*m)[n-1]
	*m = (*m)[:n-1]
	return
}

func (m *JSONArray[V]) PushFront(value V) {
	*m = append([]V{value}, *m...)
}

func (m *JSONArray[V]) PopFront() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	v = (*m)[0]
	*m = (*m)[1:]
	return
}

func (m *JSONArray[V]) Front() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	return (*m)[0], true
}

func (m *JSONArray[V]) Back() (v V, ok bool) {
	n := len(*m)
	if n == 0 {
		return v, false
	}

	return (*m)[n-1], true
}

func (m *JSONArray[V]) Get(idx int) (v V, ok bool) {
	if idx < 0 || idx >= len(*m) {
		return v, false
	}

	return (*m)[idx], true
}

func (m *JSONArray[V]) Remove(idx int) (v V, ok bool) {
	if idx < 0 || idx >= len(*m) {
		return v, false
	}

	if idx == 0 {
		return m.PopFront()
	}
	if idx == len(*m)-1 {
		return m.PopBack()
	}

	v = (*m)[idx]
	*m = append((*m)[0:idx], (*m)[idx+1:]...)

	return v, true
}

// Value return json value, implement driver.Valuer interface
func (m *JSONArray[V]) Value() (driver.Value, error) {
	return GetValue(m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *JSONArray[V]) Scan(value any) error {
	return ScanValue(value, m)
}

// GormDBDataType implements migrator.GormDBDataTypeInterface interface
func (m *JSONArray[V]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return GetGormDBDataType(db, field)
}

type Map[K comparable, V Builtin] map[K]V

// Value return json value, implement driver.Valuer interface
func (m *Map[K, V]) Value() (driver.Value, error) {
	return GetValue(m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *Map[K, V]) Scan(value any) error {
	return ScanValue(value, m)
}

// GormDBDataType implements migrator.GormDBDataTypeInterface interface
func (m *Map[K, V]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return GetGormDBDataType(db, field)
}

type JSONMap[K comparable, V JSONValue] map[K]V

// Value return json value, implement driver.Valuer interface
func (m *JSONMap[K, V]) Value() (driver.Value, error) {
	return GetValue(m)
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (m *JSONMap[K, V]) Scan(value any) error {
	return ScanValue(value, m)
}

// GormDBDataType implements migrator.GormDBDataTypeInterface interface
func (m *JSONMap[K, V]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return GetGormDBDataType(db, field)
}

// GetValue return json value, implement driver.Valuer interface
func GetValue(m any) (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	return string(b), err
}

// ScanValue scan value into Jsonb, implements sql.Scanner interface
func ScanValue(value, m any) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", value))
	}

	return json.Unmarshal(bytes, &m)
}

func GetGormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql", "sqlite", "splite3":
		return "JSON"
	case "postgres":
		return "JSONB"
	}
	return ""
}

// JSONQueryExpression json query expression, implements clause.Expression interface to use as querier
type JSONQueryExpression struct {
	column      string
	keys        []string
	hasKeys     bool
	contains    bool
	equals      bool
	equalsValue interface{}
	extract     bool
	path        string
}

// JSONQuery query column as json
func JSONQuery(column string) *JSONQueryExpression {
	return &JSONQueryExpression{column: column}
}

// Extract extract json with path
func (jsonQuery *JSONQueryExpression) Extract(path string) *JSONQueryExpression {
	jsonQuery.extract = true
	jsonQuery.path = path
	return jsonQuery
}

// HasKey returns clause.Expression
func (jsonQuery *JSONQueryExpression) HasKey(keys ...string) *JSONQueryExpression {
	jsonQuery.keys = keys
	jsonQuery.hasKeys = true
	return jsonQuery
}

// Contains returns clause.Expression
func (jsonQuery *JSONQueryExpression) Contains(tx *gorm.DB, values interface{}, keys ...string) (*JSONQueryExpression, string) {
	var query string
	switch tx.Dialector.Name() {
	case "sqlite":
		query = fmt.Sprintf("INNER JOIN JSON_EACH(%s)", tx.Statement.Quote(jsonQuery.column))
	case "postgres":
		query = fmt.Sprintf("CROSS JOIN LATERAL jsonb_array_elements(%s) o%s", tx.Statement.Quote(jsonQuery.column), jsonQuery.column)
	}
	jsonQuery.keys = keys
	jsonQuery.contains = true
	jsonQuery.equalsValue = values
	return jsonQuery, query
}

// Equals Keys returns clause.Expression
func (jsonQuery *JSONQueryExpression) Equals(value interface{}, keys ...string) *JSONQueryExpression {
	jsonQuery.keys = keys
	jsonQuery.equals = true
	jsonQuery.equalsValue = value
	return jsonQuery
}

// Build implements clause.Expression
func (jsonQuery *JSONQueryExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		switch stmt.Dialector.Name() {
		case "mysql":
			switch {
			case jsonQuery.contains:

				if len(jsonQuery.keys) == 0 {
					builder.WriteString(fmt.Sprintf("JSON_CONTAINS(%s, '%s')", stmt.Quote(jsonQuery.column), jsonQuery.equalsValue))
				} else {
					var sm string
					for i := len(jsonQuery.keys) - 1; i >= 0; i-- {
						if i == len(jsonQuery.keys)-1 {
							sm = fmt.Sprintf("JSON_OBJECT('%s', '%s')", jsonQuery.keys[i], jsonQuery.equalsValue)
						} else {
							sm = fmt.Sprintf("JSON_OBJECT('%s', %s)", jsonQuery.keys[i], sm)
						}
					}
					builder.WriteString(fmt.Sprintf("JSON_CONTAINS(%s, %s)", stmt.Quote(jsonQuery.column), sm))
				}

			case jsonQuery.extract:
				builder.WriteString("JSON_EXTRACT(")
				builder.WriteQuoted(jsonQuery.column)
				builder.WriteByte(',')
				builder.AddVar(stmt, jsonQuery.path)
				builder.WriteString(")")
			case jsonQuery.hasKeys:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString("JSON_EXTRACT(")
					builder.WriteQuoted(jsonQuery.column)
					builder.WriteByte(',')
					builder.AddVar(stmt, jsonQueryJoin(jsonQuery.keys))
					builder.WriteString(") IS NOT NULL")
				}
			case jsonQuery.equals:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString("JSON_EXTRACT(")
					builder.WriteQuoted(jsonQuery.column)
					builder.WriteByte(',')
					builder.AddVar(stmt, jsonQueryJoin(jsonQuery.keys))
					builder.WriteString(") = ")
					if value, ok := jsonQuery.equalsValue.(bool); ok {
						builder.WriteString(strconv.FormatBool(value))
					} else {
						stmt.AddVar(builder, jsonQuery.equalsValue)
					}
				}
			}

		case "sqlite":

			switch {
			case jsonQuery.contains:
				if len(jsonQuery.keys) == 0 {
					builder.WriteString(fmt.Sprintf("JSON_EACH.value %s ", "="))
				} else {
					builder.WriteString(fmt.Sprintf("JSON_EXTRACT(JSON_EACH.value, '$.%s') %s ", strings.Join(jsonQuery.keys, "."), "="))
				}
				stmt.AddVar(builder, jsonQuery.equalsValue)
			case jsonQuery.extract:
				builder.WriteString("JSON_EXTRACT(")
				builder.WriteQuoted(jsonQuery.column)
				builder.WriteByte(',')
				builder.AddVar(stmt, jsonQuery.path)
				builder.WriteString(")")
			case jsonQuery.hasKeys:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString("JSON_EXTRACT(")
					builder.WriteQuoted(jsonQuery.column)
					builder.WriteByte(',')
					builder.AddVar(stmt, jsonQueryJoin(jsonQuery.keys))
					builder.WriteString(") IS NOT NULL")
				}
			case jsonQuery.equals:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString("JSON_EXTRACT(")
					builder.WriteQuoted(jsonQuery.column)
					builder.WriteByte(',')
					builder.AddVar(stmt, jsonQueryJoin(jsonQuery.keys))
					builder.WriteString(") = ")
					if value, ok := jsonQuery.equalsValue.(bool); ok {
						builder.WriteString(strconv.FormatBool(value))
					} else {
						stmt.AddVar(builder, jsonQuery.equalsValue)
					}
				}
			}

		case "postgres":
			switch {
			case jsonQuery.contains:
				if len(jsonQuery.keys) == 0 {
					builder.WriteString(fmt.Sprintf("%s ? ", stmt.Quote("o"+jsonQuery.column)))
				} else {
					builder.WriteString(fmt.Sprintf("%s %s ", pgJoin("o"+jsonQuery.column, jsonQuery.keys...), "="))
				}
				stmt.AddVar(builder, jsonQuery.equalsValue)
			case jsonQuery.hasKeys:
				if len(jsonQuery.keys) > 0 {
					stmt.WriteQuoted(jsonQuery.column)
					stmt.WriteString("::jsonb")
					for _, key := range jsonQuery.keys[0 : len(jsonQuery.keys)-1] {
						stmt.WriteString(" -> ")
						stmt.AddVar(builder, key)
					}

					stmt.WriteString(" ? ")
					stmt.AddVar(builder, jsonQuery.keys[len(jsonQuery.keys)-1])
				}
			case jsonQuery.equals:
				if len(jsonQuery.keys) > 0 {
					builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(jsonQuery.column)))

					for idx, key := range jsonQuery.keys {
						if idx > 0 {
							builder.WriteByte(',')
						}
						stmt.AddVar(builder, key)
					}
					builder.WriteString(") = ")

					if _, ok := jsonQuery.equalsValue.(string); ok {
						stmt.AddVar(builder, jsonQuery.equalsValue)
					} else {
						stmt.AddVar(builder, fmt.Sprint(jsonQuery.equalsValue))
					}
				}
			}
		}
	}
}

const prefix = "$."

func jsonQueryJoin(keys []string) string {
	if len(keys) == 1 {
		return prefix + keys[0]
	}

	n := len(prefix)
	n += len(keys) - 1
	for i := 0; i < len(keys); i++ {
		n += len(keys[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(prefix)
	b.WriteString(keys[0])
	for _, key := range keys[1:] {
		b.WriteString(".")
		b.WriteString(key)
	}
	return b.String()
}

func pgJoin(column string, keys ...string) string {
	if len(keys) == 1 {
		return column + "->>" + "'" + keys[0] + "'"
	}
	outs := []string{column}
	for item, key := range keys {
		if item == len(keys)-1 {
			outs = append(outs, "->>")
		} else {
			outs = append(outs, "->")
		}
		outs = append(outs, "'"+key+"'")
	}
	return strings.Join(outs, "")
}
