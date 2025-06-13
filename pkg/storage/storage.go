package storage

import "github.com/cardil/qe-clusterlogging/pkg/clusterlogging"

type Storage interface {
	Store(msg *clusterlogging.Message) error
	Stats() Stats
	Download() Artifacts
}
