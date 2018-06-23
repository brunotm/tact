package linux

import (
	"time"

	"github.com/brunotm/tact/collector/keys"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/js"
)

func init() {
	tact.Registry.Add(netIOStat)
}

var netIOStat = &tact.Collector{
	Name:    "/linux/performance/netiostat",
	GetData: netIOStatFn,
	PostOps: netIOStatPostOpsFn,
	EventOps: &tact.EventOps{
		Delta: &tact.DeltaOps{
			KeyField:  keys.Device,
			TTL:       15 * time.Minute,
			Rate:      true,
			Blacklist: tact.BuildBlackList(keys.Device),
		},
	},
}

var netIOStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Device, rexon.String),
		rexon.MustNewValue(keys.NetMBRXAvg, rexon.DigitalUnit, rexon.ToFormat("mb")),
		rexon.MustNewValue(keys.NetPacketsRXAvg, rexon.Number),
		rexon.MustNewValue(keys.NetErrorsRXAvg, rexon.Number),
		rexon.MustNewValue(keys.NetDropsRXAvg, rexon.Number),
		rexon.MustNewValue(keys.NetMBTXAvg, rexon.DigitalUnit, rexon.ToFormat("mb")),
		rexon.MustNewValue(keys.NetPacketsTXAvg, rexon.Number),
		rexon.MustNewValue(keys.NetErrorsTXAvg, rexon.Number),
		rexon.MustNewValue(keys.NetDropsTXAvg, rexon.Number),
	},
	rexon.TrimSpaces(),
	rexon.LineRegex(`(.*?):\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+\d+\s+\d+\s+\d+\s+\d+\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+\d+\s+\d+\s+\d+\s+\d+`),
)

func netIOStatFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, "cat /proc/net/dev", netIOStatParser)
}

func netIOStatPostOpsFn(event []byte) (out []byte, err error) {
	// Add total averages for each rx/tx measurement
	pktRX, _ := js.GetFloat(event, keys.NetPacketsRXAvg)
	pktTX, _ := js.GetFloat(event, keys.NetPacketsTXAvg)
	event, _ = js.Set(event, pktRX+pktTX, keys.NetPacketsAvg)
	mbRX, _ := js.GetFloat(event, keys.NetMBRXAvg)
	mbTX, _ := js.GetFloat(event, keys.NetMBTXAvg)
	event, _ = js.Set(event, mbRX+mbTX, keys.NetMBAvg)
	errorsRX, _ := js.GetFloat(event, keys.NetErrorsRXAvg)
	errorsTX, _ := js.GetFloat(event, keys.NetErrorsTXAvg)
	event, _ = js.Set(event, errorsRX+errorsTX, keys.NetErrorsAvg)
	dropsRX, _ := js.GetFloat(event, keys.NetDropsRXAvg)
	dropsTX, _ := js.GetFloat(event, keys.NetDropsTXAvg)
	event, _ = js.Set(event, dropsRX+dropsTX, keys.NetDropsAvg)
	return event, nil
}
