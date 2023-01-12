// Code generated by proto-gen-validator
// source: github.com/vine-io/apimachinery/apis/meta/v1/meta.proto

package metav1

import (
	fmt "fmt"

	is "github.com/vine-io/vine/util/is"
)

func (m *TypeMeta) Validate() error {
	return m.ValidateE("")
}

func (m *TypeMeta) ValidateE(prefix string) error {
	errs := make([]error, 0)
	return is.MargeErr(errs...)
}

func (m *ObjectMeta) Validate() error {
	return m.ValidateE("")
}

func (m *ObjectMeta) ValidateE(prefix string) error {
	errs := make([]error, 0)
	if len(m.Uid) != 0 {
	}
	return is.MargeErr(errs...)
}

func (m *OwnerReference) Validate() error {
	return m.ValidateE("")
}

func (m *OwnerReference) ValidateE(prefix string) error {
	errs := make([]error, 0)
	return is.MargeErr(errs...)
}

func (m *State) Validate() error {
	return m.ValidateE("")
}

func (m *State) ValidateE(prefix string) error {
	errs := make([]error, 0)
	if int32(m.Code) != 0 {
		if !is.In([]int32{0, 1, 2}, int32(m.Code)) {
			errs = append(errs, fmt.Errorf("field '%scode' must in '[0, 1, 2]'", prefix))
		}
	}
	return is.MargeErr(errs...)
}