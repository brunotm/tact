package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/collector/keys"
)

const (
	lspvCmd = "lspv"
)

func init() {
	tact.Registry.Add(lspv)
}

var lspv = &tact.Collector{
	Name:    "/aix/config/lspv",
	GetData: lspvFn,
}

var lspvParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Device, rexon.String),
		rexon.MustNewValue("pvid", rexon.String),
		rexon.MustNewValue(keys.VGName, rexon.String),
		rexon.MustNewValue(keys.VGMode, rexon.String),
	},
	rexon.LineRegex(`(\w+)\s+(\w+)\s+(\w+)\s+(\w+)`),
)

func lspvFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, lspvCmd, lspvParser)
}
