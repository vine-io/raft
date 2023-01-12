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

package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
// concepts during lookup stages without having partially valid types
type GroupKind struct {
	Group string
	Kind  string
}

func (gk GroupKind) Empty() bool {
	return len(gk.Group) == 0 && len(gk.Kind) == 0
}

func (gk GroupKind) WithVersion(version string) GroupVersionKind {
	return GroupVersionKind{Group: gk.Group, Version: version, Kind: gk.Kind}
}

func (gk GroupKind) String() string {
	if len(gk.Group) == 0 {
		return gk.Kind
	}
	return gk.Kind + "." + gk.Group
}

// GroupVersionKind contains the information of Entity, etc Group, Version, Kind
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

// Empty returns true if group, version, and kind are empty
func (gvk GroupVersionKind) Empty() bool {
	return len(gvk.Group) == 0 && len(gvk.Version) == 0 && len(gvk.Kind) == 0
}

func (gvk GroupVersionKind) GroupKind() GroupKind {
	return GroupKind{Group: gvk.Group, Kind: gvk.Kind}
}

func (gvk GroupVersionKind) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvk.Group, Version: gvk.Version}
}

func (gvk GroupVersionKind) APIGroup() string {
	if gvk.Group == "" {
		return gvk.Version
	}
	return gvk.Group + "/" + gvk.Version
}

func (gvk GroupVersionKind) String() string {
	var s string
	if gvk.Group != "" {
		s = gvk.Group + "/"
	}
	if gvk.Version != "" {
		s = s + gvk.Version + "."
	}
	return s + gvk.Kind
}

func FromGVK(s string) GroupVersionKind {
	gvk := GroupVersionKind{}
	if idx := strings.Index(s, "/"); idx != -1 {
		gvk.Group = s[:idx]
		s = s[idx+1:]
	}
	if idx := strings.Index(s, "."); idx != -1 {
		gvk.Version = s[:idx]
		s = s[idx+1:]
	} else {
		gvk.Version = "v1"
	}
	gvk.Kind = s
	return gvk
}

// GroupVersion contains the "group" and the "version", which uniquely identifies the API.
type GroupVersion struct {
	Group   string
	Version string
}

// Empty returns true if group and version are empty
func (gv GroupVersion) Empty() bool {
	return len(gv.Group) == 0 && len(gv.Version) == 0
}

// String puts "group" and "version" into a single "group/version" string. For the legacy v1
// it returns "v1".
func (gv GroupVersion) String() string {
	if len(gv.Group) > 0 {
		return gv.Group + "/" + gv.Version
	}
	return gv.Version
}

// ParseGroupVersion turns "group/version" string into a GroupVersion struct. It reports error
// if it cannot parse the string.
func ParseGroupVersion(gv string) (GroupVersion, error) {
	// this can be the internal version for the legacy kube types
	// TODO once we've cleared the last uses as strings, this special case should be removed.
	if (len(gv) == 0) || (gv == "/") {
		return GroupVersion{}, nil
	}

	switch strings.Count(gv, "/") {
	case 0:
		return GroupVersion{"", gv}, nil
	case 1:
		i := strings.Index(gv, "/")
		return GroupVersion{gv[:i], gv[i+1:]}, nil
	default:
		return GroupVersion{}, fmt.Errorf("unexpected GroupVersion string: %v", gv)
	}
}

// WithKind creates a GroupVersionKind based on the method receiver's GroupVersion and the passed Kind.
func (gv GroupVersion) WithKind(kind string) GroupVersionKind {
	return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
}

// WithTypeKind creates a GroupVersionKind based on the method receiver's GroupVersion and the passed reflect.Type.
func (gv GroupVersion) WithTypeKind(t reflect.Type) GroupVersionKind {
	return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: t.Name()}
}

// WithAnyKind creates a GroupVersionKind based on the method receiver's GroupVersion and the passed interface.
func (gv GroupVersion) WithAnyKind(t interface{}) GroupVersionKind {
	return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: reflect.TypeOf(t).Name()}
}
