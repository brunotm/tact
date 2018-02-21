package linux

// Change to use /proc/diskstats
// https://www.kernel.org/doc/Documentation/iostats.txt

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(ioStat)
}

var ioStat = &tact.Collector{
	Name:    "/linux/performance/iostat",
	GetData: ioStatFn,
	PostOps: ioStatPostOps,
	EventOps: &tact.EventOps{
		Round: 2,
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
			"maj":            rexon.TypeString,
			"min":            rexon.TypeString,
			"device":         rexon.TypeString,
		},
		Delta: &tact.DeltaOps{
			KeyField:  "device",
			TTL:       15 * time.Minute,
			Rate:      true,
			Blacklist: tact.BuildBlackList("device", "maj", "min", "in_flight_ios"),
		},
	},
	Joins: []*tact.Join{
		{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/lsblk",
			JoinFields:    []string{"device"},
			JoinOnFields:  []string{"dm_device", "device"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn", "size_mbytes", "dm_device"},
		},
		{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/pvs",
			JoinFields:    []string{"device"},
			JoinOnFields:  []string{"dm_device", "device"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn", "size_mbytes", "vg_name", "vg_type", "vg_mode"},
		},
		{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/asm",
			JoinFields:    []string{"device"},
			JoinOnFields:  []string{"dm_device", "device"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn", "size_mbytes", "vg_name", "vg_type", "vg_mode", "asm_device"},
		},
	},
}

var ioStatParser = &rexon.RexLine{
	Rex: rexon.RexMustCompile(`(\d+)\s+(\d+)\s+(.*?)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)`),
	Fields: []string{
		"maj", "min", "device",
		"avg_reads", "avg_reads_merged", "avg_read_kb", "avg_read_svctm_ms",
		"avg_writes", "avg_writes_merged", "avg_write_kb", "avg_write_svctm_ms",
		"in_flight_ios", "avg_svctm_ms", "avg_wait_ms"},
}

func ioStatFn(session *tact.Session) <-chan []byte {
	return collector.SSHRex(session, "cat /proc/diskstats", ioStatParser)
}

func ioStatPostOps(event []byte) ([]byte, error) {

	// calculate avg_io_rate
	reads, _ := rexon.JSONGetFloat(event, "avg_reads")
	writes, _ := rexon.JSONGetFloat(event, "avg_writes")
	event, _ = rexon.JSONSet(event, rexon.Round(reads+writes, 2), "avg_io_rate")

	// maj:min
	maj, _ := rexon.JSONGetUnsafeString(event, "maj")
	min, _ := rexon.JSONGetUnsafeString(event, "min")
	event, _ = rexon.JSONSet(event, maj+":"+min, "maj_min")
	event = rexon.JSONDelete(event, "maj")
	event = rexon.JSONDelete(event, "min")

	// avg_read/write_kb
	readSect, _ := rexon.JSONGetFloat(event, "avg_read_kb")
	event, _ = rexon.JSONSet(event, (readSect*512)/1024, "avg_read_kb")
	writeSect, _ := rexon.JSONGetFloat(event, "avg_write_kb")
	event, _ = rexon.JSONSet(event, (writeSect*512)/1024, "avg_write_kb")

	// avg_wait_ms and avg_svctm_total_ms
	svctm, _ := rexon.JSONGetFloat(event, "avg_svctm_ms")
	wait, _ := rexon.JSONGetFloat(event, "avg_wait_ms")
	event, _ = rexon.JSONSet(event, rexon.Round(wait-svctm, 2), "avg_wait_ms")
	event, _ = rexon.JSONSet(event, wait, "avg_svctm_total_ms")

	return event, nil
}
