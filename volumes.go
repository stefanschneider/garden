package garden

import "time"

// A Volume is a named directory in the host with some associated properties.
type Volume interface {
	PropertyManager

	// Id returns the identifier of the volume.
	Id() string

	// Path returns the host path of the volume's directory.
	Path() string
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

type VolumeMode uint8

const (
	VolumeModeRO VolumeMode = iota
	VolumeModeRW
)

// Extension to the client interface. Manages a collection of volumes associated with the server.
type VolumeManager interface {
	// CreateVolume creates an empty volume with a TTL. Returns the new volume.
	CreateVolume(ttl time.Duration) (Volume, error)

	// CreateVolumeFromPath creates a new volume with a TTL. The volume directly accesses
	// the directory at the given path on the host. The given mode determines the mode of access
	// (note: it does not produce a copy-on-write layer). Returns the new volume.
	CreateVolumeFromPath(ttl time.Duration, hostPath string, mode VolumeMode) (Volume, error)

	// CreateVolumeFromVolume creates a new volume with a TTL. The volume is a logical copy of
	// the base volume. The given mode determines the mode of access. The properties of the new volume
	// are a snapshot of the properties of the base volume. Returns the new volume.
	// Note: read-write mode may be implemented using copy on write.
	CreateVolumeFromVolume(ttl time.Duration, baseVolume Volume, mode VolumeMode) (Volume, error)

	// GetVolume returns the Volume with the given id.
	GetVolume(id string) (Volume, error)

	// DestroyVolume destroys the volume with the given id. The volume is marked for deletion and
	// removed from the set of known volumes. Underlying system resources are not released until the
	// volume is no longer referenced.
	DestroyVolume(id string) error
}
