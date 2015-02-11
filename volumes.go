package garden

import "time"

// Note: Volumes should generate their ids and not be named.
// Pass ids instead of Volumes on other interfaces
// Consider a volume collection object

// A Volume is a named directory in the host with some associated properties.
type Volume interface {
	PropertyManager

	// Name returns the name of the volume.
	Name() string

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

// CreateVolume creates an empty volume with a given name and TTL.
func CreateVolume(name string, ttl time.Duration) (Volume, error) {
	return nil, nil
}

// CreateVolumeFromPath creates a new volume with a given name and TTL. The volume directly accesses
// the directory at the given path on the host. The given mode determines the mode of access (note: it does not
// produce a copy on write layer).
func CreateVolumeFromPath(name string, ttl time.Duration, hostPath string, mode VolumeMode) (Volume, error) {
	return nil, nil
}

// CreateVolumeFromVolume creates a new volume with a given name and TTL. The volume is a logical copy of
// the base volume. The given mode determines the mode of access. The properties of the new volume
// are a snapshot of the properties of the base volume.
// Note: read-write mode may be implemented using copy on write.
func CreateVolumeFromVolume(name string, ttl time.Duration, baseVolume Volume, mode VolumeMode) (Volume, error) {
	return nil, nil
}

// DestroyVolume destroys the volume with the given name.
func DestroyVolume(name string) error {
	return nil
}
