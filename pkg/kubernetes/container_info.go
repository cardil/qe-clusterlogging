package kubernetes

import (
	"path"
)

type ContainerInfo struct {
	ContainerName  string `json:"container_name"`
	ContainerImage string `json:"container_image"`
	NamespaceName  string `json:"namespace_name"`
	PodName        string `json:"pod_name"`
}

func (i ContainerInfo) FullName() string {
	return path.Join(i.NamespaceName, i.PodName, i.ContainerName)
}
