package collector

import (
	"log/slog"

	"github.com/cardil/qe-clusterlogging/pkg/clusterlogging"
	"github.com/cardil/qe-clusterlogging/pkg/storage"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type Collector struct {
	storage.Storage
}

func (c *Collector) Collect(channel syslog.LogPartsChannel) {
	for logParts := range channel {
		if err := c.processLog(logParts); err != nil {
			slog.Error("Log processing failed", "error", err)
			continue
		}
	}
}

func (c *Collector) processLog(logParts format.LogParts) error {
	data, err := clusterlogging.Parse(logParts)
	if err != nil {
		return err
	}
	if err = c.Store(data); err != nil {
		return err
	}
	return nil
}
