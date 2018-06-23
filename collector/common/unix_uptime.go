package common

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
)

const (
	upTimeCmd = "uptime"
)

// NewUnixUptimeFn creates a new unix uptime GetData
func NewUnixUptimeFn() tact.GetDataFn {
	return func(ctx *tact.Context) <-chan []byte {
		return ssh.Regex(ctx, upTimeCmd, upTimeParser)
	}
}

var upTimeParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue("days_up", rexon.Number, rexon.ValueRegex(`(\d+)\s+day`)),
		rexon.MustNewValue("users", rexon.Number, rexon.ValueRegex(`(\d+)\s+user`)),
		rexon.MustNewValue(
			"load_average_01",
			rexon.Number,
			rexon.ValueRegex(`load average.*?([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"load_average_05",
			rexon.Number,
			rexon.ValueRegex(`load average.*?[-+]?[0-9]*\.?[0-9]+,\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"load_average_15",
			rexon.Number,
			rexon.ValueRegex(`load average.*?[-+]?[0-9]*\.?[0-9]+,\s+[-+]?[0-9]*\.?[0-9]+,\s+([-+]?[0-9]*\.?[0-9]+)`)),
	},
)
