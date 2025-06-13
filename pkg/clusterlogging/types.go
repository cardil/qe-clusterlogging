package clusterlogging

import (
	"time"

	"github.com/cardil/qe-clusterlogging/pkg/kubernetes"
)

type Message struct {
	Timestamp                time.Time `json:"timestamp"`
	Message                  string    `json:"message"`
	kubernetes.ContainerInfo `json:"kubernetes"`
}
