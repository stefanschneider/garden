package garden

import (
	"fmt"
	"time"
)

/*
 * Example Usage:
 *
 *
 * Rootfses:
 * --------
 *
 * - Using a Docker image to create a root filesystem:
 * dockerImporter, _ := NewDockerImageImporter("url")
 * image, _ := dockerImporter.Import("ubuntu:latest")
 * rootfs, _ := image.Mount(printProgress(), 1 * time.Hour)
 *
 * - Using a host directory to create a root filesystem:
 * image, _ := CreateHostImage("path/on/host")
 * rootfs, _ := image.Mount(printProgress(), 1 * time.Hour)
 *
 * - Using a Rocket image to create a root filesystem:
 * rocketImporter, _ := NewRocketImageImporter(...)
 * image, _ := rocketImporter.Import(...)
 * rootfs, _ := image.Mount(printProgress(), 1 * time.Hour)
 *
 * - Creating a container with a rootfs:
 * container := client.Create(ContainerSpec{ Rootfs: rootfs })
 *
 *
 * Volumes:
 * -------
 *
 * - Creating a new empty Volume:
 * volume1, _ := client.CreateVolume(1 * time.Hour)
 *
 * - Creating a Volume from a host directory:
 * volume2, _ := client.CreateVolumeFromPath(1 * time.Hour, "path/on/host", VolumeModeRO)
 *
 * - Creating a Volume from another Volume:
 * volume3, _ := client.CreateVolumeFromVolume(1 * time.Hour, volume2, VolumeModeRW)
 *
 * - Binding a volume into a container:
 * ctr, _ := client.Create(ContainerSpec{
 * 				BindVolumes: []BindVolumeSpec{
 * 						Volume:     volume2,
 * 						TargetPath: "/foo",
 * 						Mode:       BindModeRW,
 * 				},
 *           })
 *
 * or, after create:
 * ctr.BindVolume(BindVolumeSpec{ Volume: volume2, TargetPath: "/foo", Mode: BindModeRW })
 *
 * - Binding a volume from one container to another
 *  (not currently supported)
 *
 *
 * Usage, reference and TTL:
 * ------------------------
 *
 * Various objects have TTLs on creation.  When an object is continuously unused for the TTL duration
 * it is marked for deletion and is deleted when no longer referenced.
 *
 * After a succesful CreateVolumeFromVolume call, the new Volume references the baseVolume. All other
 * Volumes are unreferenced.
 * A Volume is used when it is either bound to a container, or referenced by a Volume which is used.
 *
 * After a container is successfully created, its Rootfs is in use. A Rootfs is not referenced.
 * When the container using a Rootfs is destroyed, the Rootfs is no longer in use.
 *
 * A container is (momentarily) used whenever it is operated on by the garden api. A container is not referenced.
 *
 *
 * Properties:
 * ----------
 *
 * Containers, Volumes and Images each have properties. A new property set is created whenever one of
 * these objects is created. The new property set is empty, except when properties are specified in
 * a ContainerSpec, or when a Volume is created from another Volume.
 */

//go:generate counterfeiter . Client

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

	VolumeManager
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

	// Rootfs is a root file system for the container, created by Image.Mount().
	Rootfs Rootfs `json:"rootfs,omitempty"`

	// BindVolumes is a list of volumes to be bound, together with the bind parameters,
	// on container creation.
	// Note: BindVolumes do not support mounting a directory with "container origin".
	BindVolumes []BindVolumeSpec `json:"bind_volumes,omitempty"`

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

type Capacity struct {
	MemoryInBytes uint64 `json:"memory_in_bytes,omitempty"`
	DiskInBytes   uint64 `json:"disk_in_bytes,omitempty"`
	MaxContainers uint64 `json:"max_containers,omitempty"`
}

type Properties map[string]string

// A PropertyManager supports CRUD operations on string properties with string names.
type PropertyManager interface {
	// Set sets the named property to the given value. If the property was previously set,
	// its value is updated.
	Set(name, value string) error

	// Has returns true if and only if the named property exists.
	Has(name string) (bool, error)

	// Get returns the value of a named property if it exists, or the empty string otherwise.
	Get(name string) (string, error)

	// Remove deletes the named property if it exists, or does nothing otherwise.
	Remove(name string) error
}

// Rootfs represents a root file system, appropriately set up for use by a container.
type Rootfs struct {
}
