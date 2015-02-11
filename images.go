package garden

import (
	"net/url"
	"time"
)

// An Image records a collection of settable properties and allows the image to be mounted as a Rootfs.
type Image interface {
	// PropertyManager associates properties with this Image.
	PropertyManager

	// Mount constructs a Rootfs from an Image. The Rootfs is logically a writeable copy of the
	// Image's file system.
	// The given TTL controls how long the Rootfs will survive if no containers refer to it.
	Mount(pm ProgressMonitor, ttl time.Duration) (Rootfs, error)
}

// A ProgressMonitor reports progress of long-running processes.
type ProgressMonitor interface {
	// Progress sets the progress to a proportion between 0 and 1 where 1 indicates
	// the long-running process is complete.
	Progress(proportion float32)
}

// A DockerImage is an Image which also records docker image metadata.
type DockerImage interface {
	Image

	// Metadata returns the metadata associated with the Docker image.
	Metadata() *DockerMetadata
}

// DockerMetadata is the metadata of a Docker image.
type DockerMetadata struct {
	Env     []string
	Volumes []string
	// TBD
}

// DockerImageImporter creates a DockerImage from a Docker repository.
type DockerImageImporter interface {
	// Import creates a DockerImage with the given id, from this importer.
	Import(id string) (DockerImage, error)
}

// Creates a DockerImageImporter from a particular repository URL
// Note: Is endpoint sufficient? What about authentication parms, for example?
func NewDockerImageImporter(endpoint url.URL) (DockerImageImporter, error) {
	return nil, nil
}

// A RocketImage is an Image which also records Rocket image metadata.
type RocketImage interface {
	Image

	// Metadata returns the metadata associated with the Rocket image.
	Metadata() *RocketMetadata
}

// RocketMetadata is the metadata of a Rocket image.
type RocketMetadata struct {
	// Note: TBD
}

// RocketImageImporter creates a RocketImage from a Rocket repository (or some such).
type RocketImageImporter interface {
	// Import creates a RocketImage from this importer.
	// Note: parameters?
	Import() (RocketImage, error)
}

// Creates a RocketImageImporter.
func NewRocketImageImporter( /* TBD */ ) (RocketImageImporter, error) {
	return nil, nil
}

// Create a host Image based on a directory at the given path.
func CreateHostImage(path string) (Image, error) {
	return nil, nil
}
