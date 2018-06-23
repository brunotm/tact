package linux

import (
	"time"

	"github.com/brunotm/tact/js"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/collector/keys"
)

func init() {
	tact.Registry.Add(ioStat)
}

var ioStat = &tact.Collector{
	Name:    "/linux/performance/iostat",
	GetData: ioStatFn,
	PostOps: ioStatPostOps,
	EventOps: &tact.EventOps{
		Round: 2,
		Delta: &tact.DeltaOps{
			KeyField:  "device",
			TTL:       15 * time.Minute,
			Rate:      true,
			Blacklist: tact.BuildBlackList(keys.Device, keys.Maj, keys.Min, "in_flight_ios"),
		},
	},
	Joins: []*tact.Join{
		{
			TTL:          3 * time.Hour,
			Name:         "/linux/config/lsblk",
			JoinFields:   []string{keys.MajMin},
			JoinOnFields: []string{keys.MajMin},
			IncludeFields: []string{
				keys.ArrayID, keys.ArrayDevice,
				keys.DeviceWWN, keys.SizeMB, "dm_device"},
		},
		{
			TTL:          3 * time.Hour,
			Name:         "/linux/config/pvs",
			JoinFields:   []string{keys.MajMin},
			JoinOnFields: []string{keys.MajMin},
			IncludeFields: []string{
				keys.ArrayID, keys.ArrayDevice,
				keys.DeviceWWN, keys.SizeMB,
				keys.VGName, keys.VGType, keys.VGMode},
		},
		{
			TTL:          3 * time.Hour,
			Name:         "/linux/config/asm",
			JoinFields:   []string{keys.MajMin},
			JoinOnFields: []string{keys.MajMin},
			IncludeFields: []string{
				keys.ArrayID, keys.ArrayDevice,
				keys.DeviceWWN, keys.SizeMB, keys.VGName,
				keys.VGType, keys.VGMode, "asm_device"},
		},
	},
}

var ioStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Maj, rexon.String),
		rexon.MustNewValue(keys.Min, rexon.String),
		rexon.MustNewValue(keys.Device, rexon.String),
		rexon.MustNewValue(keys.IOReadRateAvg, rexon.Number),
		rexon.MustNewValue("avg_reads_merged", rexon.Number),
		rexon.MustNewValue(keys.IORateReadMBAvg, rexon.Number),
		rexon.MustNewValue(keys.IOLatencyReadMSAvg, rexon.Number),
		rexon.MustNewValue(keys.IOWriteRateAvg, rexon.Number),
		rexon.MustNewValue("avg_writes_merged", rexon.Number),
		rexon.MustNewValue(keys.IORateWriteMBAvg, rexon.Number),
		rexon.MustNewValue(keys.IOLatencyWriteMSAvg, rexon.Number),
		rexon.MustNewValue("in_flight_ios", rexon.Number),
		rexon.MustNewValue(keys.IOServiceMSAvg, rexon.Number),
		rexon.MustNewValue(keys.IOWaitMSAvg, rexon.Number),
	},
	rexon.LineRegex(`(\d+)\s+(\d+)\s+(.*?)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`),
)

func ioStatFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, "cat /proc/diskstats", ioStatParser)
}

func ioStatPostOps(event []byte) (out []byte, err error) {

	// calculate avg_io_rate
	reads, err := js.GetFloat(event, keys.IORateReadMBAvg)
	writes, err := js.GetFloat(event, keys.IORateWriteMBAvg)
	event, err = js.Set(event, reads+writes, keys.IORateAvg)

	// maj:min
	maj, err := js.GetUnsafeString(event, keys.Maj)
	min, err := js.GetUnsafeString(event, keys.Min)
	event, err = js.Set(event, maj+":"+min, keys.MajMin)
	event = js.Delete(event, keys.Maj)
	event = js.Delete(event, keys.Min)

	// avg_read/write_kb
	readSect, err := js.GetFloat(event, keys.IORateReadMBAvg)
	event, err = js.Set(event, (readSect*512)/1000, keys.IORateReadMBAvg)
	writeSect, err := js.GetFloat(event, keys.IORateWriteMBAvg)
	event, err = js.Set(event, (writeSect*512)/1000, keys.IORateWriteMBAvg)

	// avg_wait_ms and avg_svctm_total_ms
	svctm, err := js.GetFloat(event, keys.IOServiceMSAvg)
	wait, err := js.GetFloat(event, keys.IOWaitMSAvg)
	event, err = js.Set(event, wait-svctm, keys.IOWaitMSAvg)
	event, err = js.Set(event, wait, keys.IOLatencyMSAvg)

	return event, err
}
