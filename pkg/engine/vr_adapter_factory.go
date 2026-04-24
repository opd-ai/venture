//go:build !vr

package engine

// NewRuntimeHeadsetAdapter returns the default headset adapter for non-vr builds.
func NewRuntimeHeadsetAdapter() VRHeadsetAdapter {
	return NewStubHeadsetAdapter()
}

// NewRuntimeControllerAdapter returns the default controller adapter for non-vr builds.
func NewRuntimeControllerAdapter() VRControllerAdapter {
	return NewStubControllerAdapter()
}
