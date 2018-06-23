package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
)

const (
	wlmStatCmd = "wlmstat"
)

func init() {
	tact.Registry.Add(wlmStat)
}

var wlmStat = &tact.Collector{
	Name:    "/aix/performance/wlmstat",
	GetData: wlmStatFn,
}

var wlmStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue("class", rexon.String),
		rexon.MustNewValue("cpu", rexon.Number),
		rexon.MustNewValue("mem", rexon.Number),
		rexon.MustNewValue("dkio", rexon.Number),
	},
	rexon.LineRegex(`(\w+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)`),
)

// WLMStat collector
func wlmStatFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, wlmStatCmd, wlmStatParser)
}
