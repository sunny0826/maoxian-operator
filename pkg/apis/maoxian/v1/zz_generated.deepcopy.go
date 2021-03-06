// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaoxianBot) DeepCopyInto(out *MaoxianBot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaoxianBot.
func (in *MaoxianBot) DeepCopy() *MaoxianBot {
	if in == nil {
		return nil
	}
	out := new(MaoxianBot)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MaoxianBot) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaoxianBotList) DeepCopyInto(out *MaoxianBotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]MaoxianBot, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaoxianBotList.
func (in *MaoxianBotList) DeepCopy() *MaoxianBotList {
	if in == nil {
		return nil
	}
	out := new(MaoxianBotList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MaoxianBotList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaoxianBotSpec) DeepCopyInto(out *MaoxianBotSpec) {
	*out = *in
	if in.RepoList != nil {
		in, out := &in.RepoList, &out.RepoList
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaoxianBotSpec.
func (in *MaoxianBotSpec) DeepCopy() *MaoxianBotSpec {
	if in == nil {
		return nil
	}
	out := new(MaoxianBotSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaoxianBotStatus) DeepCopyInto(out *MaoxianBotStatus) {
	*out = *in
	if in.RepoStatus != nil {
		in, out := &in.RepoStatus, &out.RepoStatus
		*out = make([]RepoStatus, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaoxianBotStatus.
func (in *MaoxianBotStatus) DeepCopy() *MaoxianBotStatus {
	if in == nil {
		return nil
	}
	out := new(MaoxianBotStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RepoStatus) DeepCopyInto(out *RepoStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RepoStatus.
func (in *RepoStatus) DeepCopy() *RepoStatus {
	if in == nil {
		return nil
	}
	out := new(RepoStatus)
	in.DeepCopyInto(out)
	return out
}
