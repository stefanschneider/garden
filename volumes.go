package garden

import "time"

// A Volume is a named directory in the host with some associated properties.
type Volume interface {
	PropertyManager

	// Handle returns the handle of the volume.
	Handle() string
}

// BindVolumeSpec collects the parameters used to bind a volume.
type BindVolumeSpec struct {
	// The volume to bind.
	Volume Volume `json:"volume,omitempty"`

	// The target path to which the volume is bound.
	TargetPath string `json:"target_path,omitempty"`

	// The mode with which the volume is bound.
	Mode BindMode `json:"mode,omitempty"`
}

type BindMode uint8

const (
	BindModeRO BindMode = iota
	BindModeRW
)

// A BoundVolume represents the binding of a volume to a container.
type BoundVolume interface {
	// Spec returns the BindVolumeSpec used to create this bound volume.
	Spec() *BindVolumeSpec
}

// Extension to the client interface. Manages a collection of volumes associated with the server.
type VolumeManager interface {
	// CreateVolume creates an empty volume with a TTL. Returns the new volume.
	// The properties of the new volume are empty.
	// If both HostPath and BaseVolume are specified, the Volume is initially empty.
	// If both HostPath and BaseVolume are specified, an error is returned.
	CreateVolume(VolumeSpec) (Volume, error)

	// GetVolume returns the Volume with the given handle.
	GetVolume(handle string) (Volume, error)

	// Volumes returns the volumes which match the given properties.
	Volumes(Properties) ([]Volume, error)

	// DestroyVolume destroys the volume with the given id. The volume is marked for deletion and
	// removed from the set of known volumes. Underlying system resources are not released until the
	// volume is no longer referenced.
	DestroyVolume(handle string) error
}

// A TTL specifies the duration that a particular object can remain unused before it is automatically deleted.
type TTL time.Duration

type VolumeSpec struct {
	// Handle, if specified, is used to refer to the
	// Volume in future requests. If it is not specified,
	// garden uses its internal Volume ID as the Volume handle.
	// TODO: are there any restrictions?
	Handle string `json:"handle,omitempty"`

	// A TTL for the Volume.
	TTL TTL

	// If specified, this volume is logically a writable copy of the given BaseVolume.
	// Writes to this volume do not affect the underlying BaseVolume.
	BaseVolume Volume

	// If specified, this volume directly accesses the directory at the given path on the host.
	// Writes to this volume affect the underlying HostPath.
	HostPath string
}
