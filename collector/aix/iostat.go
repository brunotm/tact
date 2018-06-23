package aix

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/collector/keys"
)

const (
	ioStatCmd = "/usr/bin/iostat -DRVl 60 1"
)

func init() {
	tact.Registry.Add(ioStat)
}

var ioStat = &tact.Collector{
	Name:    "/aix/performance/iostat",
	GetData: ioStatFn,
	Joins: []*tact.Join{
		{
			TTL:          3 * time.Hour,
			Name:         "/aix/config/storage",
			JoinFields:   []string{keys.Device},
			JoinOnFields: []string{keys.Device},
			IncludeFields: []string{keys.ArrayID,
				keys.ArrayDevice, keys.DeviceWWN},
		},
		{
			TTL:           3 * time.Hour,
			Name:          "/aix/config/lspv",
			JoinFields:    []string{keys.Device},
			JoinOnFields:  []string{keys.Device},
			IncludeFields: []string{keys.VGName, keys.VGMode, "pvid"},
		},
	},
}

// Example raw data
// Disks:                     xfers                                read                                write                                  queue
// -------------- -------------------------------- ------------------------------------ ------------------------------------ --------------------------------------
//                  %tm    bps   tps  bread  bwrtn   rps    avg    min    max time fail   wps    avg    min    max time fail    avg    min    max   avg   avg  serv
//                  act                                    serv   serv   serv outs              serv   serv   serv outs        time   time   time  wqsz  sqsz qfull
// hdisk1414        0.0  34.8K   2.5   8.2K  26.6K   0.5   0.4    0.4    0.4     0    0   2.0   0.7    0.6    0.7     0    0   0.0    0.0    0.0    0.0   0.0   0.0

var ioStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Device, rexon.String),
		rexon.MustNewValue("tm_act_avg_pct", rexon.Number),
		rexon.MustNewValue(keys.IORateMBAvg, rexon.DigitalUnit, rexon.ToFormat("mb")),
		rexon.MustNewValue(keys.IORateAvg, rexon.Number),
		rexon.MustNewValue(keys.IORateReadMBAvg, rexon.DigitalUnit, rexon.ToFormat("mb")),
		rexon.MustNewValue(keys.IORateWriteMBAvg, rexon.DigitalUnit, rexon.ToFormat("mb")),
		rexon.MustNewValue(keys.IOReadRateAvg, rexon.Number),
		rexon.MustNewValue(keys.IOServiceReadMSAvg, rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("read_min_service_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("read_max_service_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("read_timeouts_avg", rexon.Number),
		rexon.MustNewValue("read_fail_avg", rexon.Number),
		rexon.MustNewValue(keys.IOWriteRateAvg, rexon.Number),
		rexon.MustNewValue(keys.IOServiceWriteMSAvg, rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("write_min_service_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("write_max_service_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("write_timeouts_avg", rexon.Number),
		rexon.MustNewValue("write_fail_avg", rexon.Number),
		rexon.MustNewValue("queue_time_avg_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("queue_time_min_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("queue_time_max_ms_avg", rexon.Duration, rexon.ToFormat("ms"), rexon.FromFormat("ms")),
		rexon.MustNewValue("queue_wait_avg_size", rexon.Number),
		rexon.MustNewValue("queue_service_avg_size", rexon.Number),
		rexon.MustNewValue("queue_service_full", rexon.Number),
	},
	rexon.LineRegex(`(hdisk\d+|[-+]?[0-9]*\.?[0-9]+\w?)`), rexon.FindAll(),
)

// iostat collector
func ioStatFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, ioStatCmd, ioStatParser)
}
