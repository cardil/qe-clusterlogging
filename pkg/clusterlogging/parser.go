package clusterlogging

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mcuadros/go-syslog.v2/format"
)

func Parse(logParts format.LogParts) (*Message, error) {
	data := &Message{}
	if err := json.Unmarshal([]byte(fmt.Sprint(logParts["message"])), data); err != nil {
		return nil, err
	}
	return data, nil
}
