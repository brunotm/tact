package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	wlmStatCmd = "wlmstat"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(wlmStat)
}

var wlmStat = &tact.Collector{
	Name:    "/aix/performance/wlmstat",
	GetData: wlmStatFn,
}

var wlmStatParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^(\w+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)`),
	Fields: []string{"class", "cpu", "mem", "dkio"},
	Round:  2,
	Types: map[string]rexon.ValueType{
		rexon.KeyTypeAll: rexon.TypeNumber,
		"class":          rexon.TypeString,
	},
}

// WLMStat collector
func wlmStatFn(session *tact.Session) <-chan []byte {
	return collector.SSHRex(session, wlmStatCmd, wlmStatParser)
}
