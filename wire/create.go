package wire

import (
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/garden"
)

type CreateRequest struct {
	BindMounts []*CreateRequest_BindMount `json:"bind_mounts",omitempty"`
	GraceTime  *uint32                    `json:"grace_time",omitempty"`
	Handle     *string                    `json:"handle",omitempty"`
	Network    *string                    `json:"network",omitempty"`
	Rootfs     *string                    `json:"rootfs",omitempty"`
	Properties []*Property                `json:"properties",omitempty"`
	Env        []*EnvironmentVariable     `json:"env",omitempty"`
	Privileged *bool                      `json:"privileged",omitempty"`
}

func NewCreateRequest(spec garden.ContainerSpec) *CreateRequest {
	return &CreateRequest{
		Handle:     OptString(spec.Handle),
		Rootfs:     OptString(spec.RootFSPath),
		GraceTime:  OptTimeSecs(spec.GraceTime),
		Network:    OptString(spec.Network),
		Env:        ConvertEnvironmentVariables(spec.Env),
		Privileged: PBool(spec.Privileged),
		BindMounts: ConvertBindMounts(spec.BindMounts),
		Properties: ConvertProperties(spec.Properties),
	}
}

type CreateRequest_BindMount struct {
	SrcPath *string                 `json:"src_path,omitempty"`
	DstPath *string                 `json:"dst_path,omitempty"`
	Mode    *garden.BindMountMode   `json:"mode,omitempty"`
	Origin  *garden.BindMountOrigin `json:"origin,omitempty"`
}

type Property struct {
	Key   *string `json:"Key,omitempty"`
	Value *string `json:"Value,omitempty"`
}

type EnvironmentVariable struct {
	Key   *string `json:"Key,omitempty"`
	Value *string `json:"Value,omitempty"`
}

type CreateResponse struct {
	Handle *string `json:"handle,omitempty"`
}

func ConvertEnvironmentVariables(environmentVariables []string) []*EnvironmentVariable {
	var convertedEnvironmentVariables []*EnvironmentVariable

	for _, env := range environmentVariables {
		segs := strings.SplitN(env, "=", 2)

		convertedEnvironmentVariables = append(
			convertedEnvironmentVariables,
			&EnvironmentVariable{
				Key:   &segs[0],
				Value: &segs[1],
			},
		)
	}
	return convertedEnvironmentVariables
}

func ConvertBindMounts(srcBms []garden.BindMount) []*CreateRequest_BindMount {
	var tgtBms []*CreateRequest_BindMount
	for _, bm := range srcBms {
		tgtBms = append(tgtBms, &CreateRequest_BindMount{
			SrcPath: PString(bm.SrcPath),
			DstPath: PString(bm.DstPath),
			Mode:    pBindMountMode(bm.Mode),
			Origin:  pBindMountOrigin(bm.Origin),
		})
	}
	return tgtBms
}

// pBindMountMode copies a bind mount mode and returns the address of the copy.
func pBindMountMode(bmm garden.BindMountMode) *garden.BindMountMode {
	return &bmm
}

// pBindMountOrigin copies a bind mount origin and returns the address of the copy.
func pBindMountOrigin(bmo garden.BindMountOrigin) *garden.BindMountOrigin {
	return &bmo
}

func ConvertProperties(srcProps garden.Properties) []*Property {
	var tgtProps []*Property
	for key, val := range srcProps {
		tgtProps = append(tgtProps, &Property{
			Key:   PString(key),
			Value: PString(val),
		})
	}
	return tgtProps
}

func OptTimeSecs(tm time.Duration) *uint32 {
	if tm == 0 {
		return nil
	}
	return pUint32(uint32(tm.Seconds()))
}

// pUint32 copies a uint32 and returns the address of the copy.
func pUint32(i uint32) *uint32 {
	return &i
}
