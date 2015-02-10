package garden

import (
	"fmt"
	"net/url"
	"time"
)

//go:generate counterfeiter . Client

type ProgressMonitor interface{}

/*
* Example Usage:
*
*
* Rootfses:
* --------
*
* - Importing a Docker RootFS:
* dockerImage := NewDockerImageImporter("url").Import("ubuntu:latest")
* rootfs := Must(dockerImage.Mount(printProgress(), 1 * time.Hour))
*
* - Importing a Garden RootFS (creates a RW layer, because you always want this for a rootfs so no need to do two steps):
* rootfs := Must(NewGardenRootfs("path/on/host", 1 * time.Hour))
*
* - Importing a Rocket RootFS (rocket!!!)
* rocketImage := NewRocketImageImporter("url").Import("foo/bar/baz")
* rootfs := Must(rocketImage.Mount(printProgress(), 1 * time.Hour)
*
* - Creating a container from a Docker rootfs:
* (nb: the TTL is handled by the Docker/Garden/RootFS backends, garden calls the Bind/Unbind methods
*  on the Volume object to inform the rootfs TTL trackers about these events, this is necessary because
*  you may never get to .Create but you still want your image to be cleaned up)
* container := client.Create(ContainerSpec{ Rootfs: rootfs })
*
*
* Volumes:
* -------
*
* - Creating a Volume:
* volume := NewHostPathVolume("path/on/host")
*
* - Importing a Volume from a HTTP address (if for some reason we think this is Garden's job):
* volume, err := ImportHTTPVolume("address", 1 * time.Hour)
*
* - Binding a volume into a container
* (note no need to expose creating read-only and read-write layers, just bind RO/RW as needed,
*  by creating the layer when we know what container it will be mounted in to we can make sure we choose
*  the right sort of filesystem overlay technique).
* container.Create(ContainerSpec{ BindMounts: []BindMount{ Volume: rwVolume, Path: "/foo", Mode: RW } })
* //or
* container.BindVolume(rwVolume, "/foo", RO)
*
* - Binding a volume from one container to another
* (below we create a volume whose source is a container - a la docker - and give it a TTL, this volume prevents
* auto-cleanup of otherContainer until it is deleted/cleaned up; this is - I think - the only multi-step dependency here)
* container.BindVolume( NewGuestVolume(otherContainer, "/foo", 1 * time.Hour), "/bar", RW)
*
* TTLs:
* ----
*
* Various objects have TTLs, this is all handled by a TTLTracker in these structs. The TTLTracker has
* Use(handle) and Unuse(handle) methods which are called by garden when a Volume/Rootfs/etc are used/mounted/unmounted.
* This allows the cleanup etc. to happen in the Rootfs/Volume components, not in garden, while not requiring the client
* to worry about that stuff. For composite TTLs (i.e. this volume depends on this volume) we could use a CompositeTTLTracker
* that knows to use/unuse its dependencies, but I *think* the above design avoids any transitive dependencies anyway (apart from
* container->rootfs and container->volume and the unavoidable transitive dependency if you mount a volume from inside a container).
*
*
* Properties:
* ----------
*
* Since DockerImage (etc), DockerRootfs, HostPathRootfs, HostVolume, BoundVolume etc. are fully
* separate objects above there is no ambiguity about their properties -- each has its own independent PropertyManager.
 */

// DockerRootfs imports a docker image from a repository and returns the metadata
type DockerImageImporter interface {
	Import(dockerId string) (DockerImage, error)
	GenerateFromDockerfile(dockerFile string) (DockerImage, error) // for funsies
}

// DockerImage is the metadata of a docker image, calling .Mount() provides a RootFS that
// can be used to launch a docker container. DockerImage has nothing to do with Rootfses at all except that it
// provides a .Mount() method to create one.
type DockerImage struct {
	Env     []string
	Volumes []string
	// etc
}

// Mount mounts a Docker Image as a Rootfs. The given TTL controls how long the mounted
// volume will survive if no containers refer to it.
func (DockerImage) Mount(pm ProgressMonitor, TTL time.Duration) (Rootfs, error) {
	// download with progress monitor, mount as a layer, return DockerVolume
	// create TTLTracker with TTLTracker.New(TTL)
	return nil, nil
}

type DockerRootfs struct {
	HostPath
	PropertyManager
	TTLTracker
}

type GardenRootfs struct {
	HostPath
	PropertyManager
	TTLTracker
}

type HostPath string

func (h HostPath) HostPath() string {
	return string(h)
}

// TTLTracker is implemented by things that have a TTL. Garden calls Use and Unuse when a particular container starts and
// stops referring to an object. The object can use this to self-destruct when noone is referring to it. Concrete implementations
// of TTLTracker can take all the work out of implementing this (obv).
type TTLTracker interface {
	Use(handle string)   // I started using you, stop the timer
	Unuse(handle string) // I stopped using you, restart the timer
}

// A Rootfs is implemented by objects which are valid values for the Rootfs field in ContainerSpec.
type Rootfs interface {
	HostPath() string // where do we mount from?
	TTLTracker        // garden will call Use() and Unuse() when a container is created/destroyed from this rootfs
}

// A HostVolume is a volume which exists on the host
type HostVolume struct {
	HostPath string
	PropertyManager
	TTLTracker
}

// A GuestVolume is a volume mounted from inside a container
type GuestVolume struct {
	ContainerHandle string
	ContainerPath   string
	PropertyManager
}

type BoundVolume struct {
	ContainerHandle string
	ContainerPath   string
	PropertyManager
}

func (BoundVolume) Unbind() error {
	return nil
}

// Creates a DockerImageImporter from a particular repository URL
func NewDockerImageImporter(endpoint url.URL) (DockerImageImporter, error) {
	return nil, nil
}

// Create a GardenRootfs by mounting a RW layer on top of a given HostPath. The rootfs is automatically
// removed once the TTL is expired unless the rootfs is used as the Rootfs of a container (in which case garden
// will call `Use` and `Unuse` appropriately so that the TTL is maintained.
func NewGardenRootfs(path HostPath, ttl time.Duration) (*GardenRootfs, error) {
	return nil, nil
}

// create a volume from inside a running container, the container's TTL will be `Used` until this
// volume times out or is destroyed, at which point it will be `Unused`.
func NewGuestVolume(container Container, path string, ttl time.Duration) (*GuestVolume, error) {
	return nil, nil
}

type PropertyManager interface{}

type Client interface {
	// Pings the garden server.
	//
	// Errors:
	// * None.
	Ping() error

	// Capacity returns the physical capacity of the server's machine.
	//
	// Errors:
	// * None.
	Capacity() (Capacity, error)

	// Create creates a new container.
	//
	// Errors:
	// * When the handle, if specified, is already taken.
	// * When one of the bind_mount paths does not exist.
	// * When resource allocations fail (subnet, user ID, etc).
	Create(ContainerSpec) (Container, error)

	// Destroy destroys a container.
	//
	// When a container is destroyed, its resource allocations are released,
	// its filesystem is removed, and all references to its handle are removed.
	//
	// All resources that have been acquired during the lifetime of the container are released.
	// Examples of these resources are its subnet, its UID, and ports that were redirected to the container.
	//
	// TODO: list the resources that can be acquired during the lifetime of a container.
	//
	// Errors:
	// * TODO.
	Destroy(handle string) error

	// Containers lists all containers filtered by Properties (which are ANDed together).
	//
	// Errors:
	// * None.
	Containers(Properties) ([]Container, error)

	// Lookup returns the container with the specified handle.
	//
	// Errors:
	// * Container not found.
	Lookup(handle string) (Container, error)
}

type ContainerNotFoundError struct {
	Handle string
}

func (err ContainerNotFoundError) Error() string {
	return fmt.Sprintf("unknown handle: %s", err.Handle)
}

// ContainerSpec specifies the parameters for creating a container. All parameters are optional.
type ContainerSpec struct {

	// Handle, if specified, is used to refer to the
	// container in future requests. If it is not specified,
	// garden uses its internal container ID as the container handle.
	Handle string `json:"handle,omitempty"`

	// GraceTime can be used to specify how long a container can go
	// unreferenced by any client connection. After this time, the container will
	// automatically be destroyed. If not specified, the container will be
	// subject to the globally configured grace time.
	GraceTime time.Duration `json:"grace_time,omitempty"`

	// RootFSPath is a URI referring to the root file system for the container.
	// The URI scheme must either be the empty string or "docker".
	//
	// A URI with an empty scheme determines the path of a root file system.
	// If this path is empty, a default root file system is used.
	// Other parts of the URI are ignored.
	//
	// A URI with scheme "docker" refers to a Docker image. The path in the URI
	// (without the leading /) identifies a Docker image as the repository name
	// in the default Docker registry. If a fragment is specified in the URI, this
	// determines the tag associated with the image.
	// If a host is specified in the URI, this determines the Docker registry to use.
	// If no host is specified in the URI, a default Docker registry is used.
	// Other parts of the URI are ignored.
	//
	// Examples:
	// * "/some/path"
	// * "docker:///onsi/grace-busybox"
	// * "docker://index.docker.io/busybox"
	RootFSPath string `json:"rootfs,omitempty"`

	// * bind_mounts: a list of mount point descriptions which will result in corresponding mount
	// points being created in the container's file system.
	//
	// An error is returned if:
	// * one or more of the mount points has a non-existent source directory, or
	// * one or more of the mount points cannot be created.
	BindMounts []BindMount `json:"bind_mounts,omitempty"`

	// Network determines the subnet and IP address of a container.
	//
	// If not specified, a /30 subnet is allocated from a default network pool.
	//
	// If specified, it takes the form a.b.c.d/n where a.b.c.d is an IP address and n is the number of
	// bits in the network prefix. a.b.c.d masked by the first n bits is the network address of a subnet
	// called the subnet address. If the remaining bits are zero (i.e. a.b.c.d *is* the subnet address),
	// the container is allocated an unused IP address from the subnet. Otherwise, the container is given
	// the IP address a.b.c.d.
	//
	// The container IP address cannot be the subnet address or the broadcast address of the subnet
	// (all non prefix bits set) or the address one less than the broadcast address (which is reserved).
	//
	// Multiple containers may share a subnet by passing the same subnet address on the corresponding
	// create calls. Containers on the same subnet can communicate with each other over IP
	// without restriction. In particular, they are not affected by packet filtering.
	//
	// Note that a container can use TCP, UDP, and ICMP, although its external access is governed
	// by filters (see Container.NetOut()) and by any implementation-specific filters.
	//
	// An error is returned if:
	// * the IP address cannot be allocated or is already in use,
	// * the subnet specified overlaps the default network pool, or
	// * the subnet specified overlaps (but does not equal) a subnet that has
	//   already had a container allocated from it.
	Network string `json:"network,omitempty"`

	// Properties is a sequence of string key/value pairs providing arbitrary
	// data about the container. The keys are assumed to be unique but this is not
	// enforced via the protocol.
	Properties Properties `json:"properties,omitempty"`

	// TODO
	Env []string `json:"env,omitempty"`

	// If Privileged is true the container does not have a user namespace and the root user in the container
	// is the same as the root user in the host. Otherwise, the container has a user namespace and the root
	// user in the container is mapped to a non-root user in the host. Defaults to false.
	Privileged bool `json:"privileged,omitempty"`
}

// BindMount specifies parameters for a single mount point.
//
// Each mount point is mounted (with the bind option) into the container's file system.
// The effective permissions of the mount point are the permissions of the source directory if the mode
// is read-write and the permissions of the source directory with the write bits turned off if the mode
// of the mount point is read-only.
type BindMount struct {
	// SrcPath contains the path of the directory to be mounted.
	SrcPath string `json:"src_path,omitempty"`

	// DstPath contains the path of the mount point in the container. If the
	// directory does not exist, it is created.
	DstPath string `json:"dst_path,omitempty"`

	// Mode must be either "RO" or "RW". Alternatively, mode may be omitted and defaults to RO.
	// If mode is "RO", a read-only mount point is created.
	// If mode is "RW", a read-write mount point is created.
	Mode BindMountMode `json:"mode,omitempty"`

	// BindMountOrigin must be either "Host" or "Container". Alternatively, origin may be omitted and
	// defaults to "Host".
	// If origin is "Host", src_path denotes a path in the host.
	// If origin is "Container", src_path denotes a path in the container.
	Origin BindMountOrigin `json:"origin,omitempty"`
}

type Capacity struct {
	MemoryInBytes uint64 `json:"memory_in_bytes,omitempty"`
	DiskInBytes   uint64 `json:"disk_in_bytes,omitempty"`
	MaxContainers uint64 `json:"max_containers,omitempty"`
}

type Properties map[string]string

type BindMountMode uint8

const BindMountModeRO BindMountMode = 0
const BindMountModeRW BindMountMode = 1

type BindMountOrigin uint8

const BindMountOriginHost BindMountOrigin = 0
const BindMountOriginContainer BindMountOrigin = 1
