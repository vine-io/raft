// Code generated by proto-gen-deepcopy. DO NOT EDIT.
// source: github.com/vine-io/vine/lib/api/api.proto

package api

import (
	registry "github.com/vine-io/vine/core/registry"
)

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
	if in.Host != nil {
		in, out := &in.Host, &out.Host
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Method != nil {
		in, out := &in.Method, &out.Method
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *Event) DeepCopyInto(out *Event) {
	*out = *in
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = make(map[string]*Pair, len(*in))
		for key, val := range *in {
			var outVal *Pair
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Pair)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *FileDesc) DeepCopyInto(out *FileDesc) {
	*out = *in
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *FileHeader) DeepCopyInto(out *FileHeader) {
	*out = *in
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *Pair) DeepCopyInto(out *Pair) {
	*out = *in
	if in.Values != nil {
		in, out := &in.Values, &out.Values
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *Request) DeepCopyInto(out *Request) {
	*out = *in
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = make(map[string]*Pair, len(*in))
		for key, val := range *in {
			var outVal *Pair
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Pair)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.Get != nil {
		in, out := &in.Get, &out.Get
		*out = make(map[string]*Pair, len(*in))
		for key, val := range *in {
			var outVal *Pair
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Pair)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.Post != nil {
		in, out := &in.Post, &out.Post
		*out = make(map[string]*Pair, len(*in))
		for key, val := range *in {
			var outVal *Pair
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Pair)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *Response) DeepCopyInto(out *Response) {
	*out = *in
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = make(map[string]*Pair, len(*in))
		for key, val := range *in {
			var outVal *Pair
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Pair)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopyInto is an auto-generated deepcopy function, coping the receiver, writing into out. in must be no-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	if in.Endpoint != nil {
		in, out := &in.Endpoint, &out.Endpoint
		*out = new(Endpoint)
		(*in).DeepCopyInto(*out)
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]*registry.Service, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(registry.Service)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}