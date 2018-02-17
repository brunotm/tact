package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	lspvCmd = "lspv"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(lspv)
}

var lspv = &tact.Collector{
	Name:    "/aix/config/lspv",
	GetData: lspvFn,
}

var lspvParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^(\w+)\s+(\w+)\s+(\w+)\s+(\w+)`),
	Fields: []string{"device", "pvid", "vg_name", "vg_mode"},
	Types:  map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeString},
}

func lspvFn(session *tact.Session) <-chan []byte {
	return collector.SSHRex(session, lspvCmd, lspvParser)
}
