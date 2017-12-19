package linux

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	multipathCmd = "/sbin/multipath -ll"
	procpartCmd  = "cat /proc/partitions"
)

// init add this collector with the registry
func init() {
	tact.Registry.Add(lsblk)
	tact.Registry.Add(pvs)
	tact.Registry.Add(asmDevices)
}

var lsblk = &tact.Collector{
	Name:    "/linux/config/lsblk",
	GetData: lsblkFn,
	Joins: []*tact.Join{
		&tact.Join{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/storage",
			JoinFields:    []string{"device", "dm_device"},
			JoinOnFields:  []string{"device"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn"},
		},
	},
}

var lsblkParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^(.*?)\s+\(?(.*?)\)?\s+(\d+:\d+)\s+\d+\s+(\d+)\s+\d+(.*)`),
	Fields: []string{"device", "dm_device", "maj_min", "size_bytes", "mount"},
	Types:  map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeString, "size_bytes": rexon.TypeNumber},
}

func lsblkFn(session tact.Session) <-chan []byte {
	outChan := make(chan []byte)
	go func() {
		defer close(outChan)
		for blk := range collector.SSHRex(session, "lsblk -lb", lsblkParser) {
			sz, err := rexon.JSONGetFloat(blk, "size_bytes")
			if err == nil {
				blk, _ = rexon.JSONSet(blk, sz/1024/1024, "size_mbytes")
				blk = rexon.JSONDelete(blk, "size_bytes")
			}

			if !tact.WrapCtxSend(session.Context(), outChan, blk) {
				session.LogErr("timeout sending event upstream")
				return
			}

		}
	}()
	return outChan
}

var pvs = &tact.Collector{
	Name:    "/linux/config/pvs",
	GetData: pvsFn,
	Joins: []*tact.Join{
		&tact.Join{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/lsblk",
			JoinFields:    []string{"device"},
			JoinOnFields:  []string{"device"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn", "size_mbytes", "device", "dm_device"},
		},
	},
}

var pvsParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^/.*?/(\w+)\s+(\w+)\s+(\w+)`),
	Fields: []string{"device", "vg_name", "vg_type"},
	Types:  map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeString},
}

func pvsFn(session tact.Session) <-chan []byte {
	return collector.SSHRex(session, "pvs", pvsParser)
}

var asmDevices = &tact.Collector{
	Name:    "/linux/config/asm",
	GetData: asmDevicesFn,
	PostOps: asmDevicesPostOpsFn,
	Joins: []*tact.Join{
		&tact.Join{
			TTL:           3 * time.Hour,
			Name:          "/linux/config/lsblk",
			JoinFields:    []string{"maj_min"},
			JoinOnFields:  []string{"maj_min"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn", "size_mbytes", "device", "dm_device"},
		},
	},
}

var asmDevicesParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^.*?\s+.*?\s+.*?\s+.*?(\d+),\s+(\d+)\s+.*?\s+.*?\s+.*?\s+.*?(.*)`),
	Fields: []string{"major", "minor", "asm_device"},
	Types:  map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeString},
}

func asmDevicesFn(session tact.Session) <-chan []byte {
	return collector.SSHRex(session, "ls -l /dev/oracleasm/disks", asmDevicesParser)
}

func asmDevicesPostOpsFn(event []byte) ([]byte, error) {
	major, _ := rexon.JSONGetString(event, "major")
	minor, _ := rexon.JSONGetString(event, "minor")
	event = rexon.JSONDelete(event, "major")
	event = rexon.JSONDelete(event, "minor")
	event, _ = rexon.JSONSet(event, "oracleasm", "vg_type")
	event, _ = rexon.JSONSet(event, nil, "vg_name")
	event, _ = rexon.JSONSet(event, nil, "vg_mode")
	return rexon.JSONSet(event, major+":"+minor, "maj_min")
}
