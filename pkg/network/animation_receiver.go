package network

// AnimationComponentType is the reserved ComponentData.Type value used to carry
// a serialized AnimationStatePacket inside a StateUpdate broadcast. The leading
// underscore prevents collision with gameplay ECS component type names.
const AnimationComponentType = "_animation"

// AnimationReceiver is implemented by network clients that accept a wired
// AnimationSyncManager for jitter-buffered remote animation state delivery.
// *TCPClient implements this; type-assert to wire after creating the manager.
type AnimationReceiver interface {
	SetAnimationSyncManager(mgr *AnimationSyncManager)
}
