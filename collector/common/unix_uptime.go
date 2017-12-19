package common

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	upTimeCmd = "uptime"
)

// NewUnixUptimeFn creates a new unix uptime GetData
func NewUnixUptimeFn() tact.GetDataFn {
	return func(session tact.Session) <-chan []byte {
		return collector.SSHRex(session, upTimeCmd, upTimeParser)
	}
}

var upTimeParser = &rexon.RexSet{
	Prep: rexon.RexMustCompile(`[(),!]`),
	Set: rexon.RexSetMustCompile(map[string]string{
		rexon.KeyStartTag: `^\d+:\d+:?\d+?\w+\s+`,
		"days_up":         `^\d+:\d+:?\d+?\w+\s+\w+\s+(\d+)\s+`,
		"users":           `^\d+:\d+:?\d+?\w+\s+\w+\s+\d+\s+\w+\s+\d+:\d+\s+(\d+)\s+`,
		"load_average_01": `^\d+:\d+:?\d+?\w+\s+.*:\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"load_average_05": `^\d+:\d+:?\d+?\w+\s+.*:\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"load_average_15": `^\d+:\d+:?\d+?\w+\s+.*:\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`}),
	Round: 2,
	Types: map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeNumber},
}
