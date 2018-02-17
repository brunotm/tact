package linux

// Change to use /proc/diskstats
// https://www.kernel.org/doc/Documentation/iostats.txt

import (
	"bytes"
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(ioStat)
}

const (
	ioStatCmd = "/usr/bin/iostat -dxk 60 2"
)

var (
	ioStatFields12 = []string{"device", "rrqm", "wrqm", "rps", "wps",
		"kbread", "kbwrtn", "avgrqsz", "avgsqusz", "await", "svctm", "util"}
	ioStatFields14 = []string{"device", "rrqm", "wrqm", "rps", "wps",
		"kbread", "kbwrtn", "avgrqsz", "avgqusz", "await", "r_await", "w_await", "svctm", "util"}
)

var ioStat = &tact.Collector{
	Name:    "/linux/performance/iostat",
	GetData: ioStatFn,
	EventOps: &tact.EventOps{
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
			"device":         rexon.TypeString,
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

func ioStatFn(session *tact.Session) <-chan []byte {
	outChan := make(chan []byte)

	go func() {
		defer close(outChan)

		scanner := collector.SSHScanner(session, ioStatCmd)
		if scanner == nil {
			return
		}

		// the current metric snapshot
		window := 0

		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				session.LogErr("error reading data: %s", err.Error())
				return
			}
			line := bytes.TrimSpace(scanner.Bytes())

			// Check and increment window
			if bytes.Contains(line, []byte("Device")) {
				if window < 2 {
					window++
					continue
				}
			}

			// Only proceed if we're in the correct window
			if window < 2 {
				continue
			}

			values := bytes.Fields(line)
			if len(values) == 12 {
				var event []byte
				for x := 0; x <= 11; x++ {
					event, _ = rexon.JSONSet(event, values[x], ioStatFields12[x])
				}
				if !tact.WrapCtxSend(session.Context(), outChan, event) {
					session.LogErr("timeout sending event upstream")
					return
				}
			}

			if len(values) == 14 {
				var event []byte
				for x := 0; x <= 13; x++ {
					event, _ = rexon.JSONSet(event, values[x], ioStatFields14[x])
				}
				if !tact.WrapCtxSend(session.Context(), outChan, event) {
					session.LogErr("timeout sending event upstream")
					return
				}
			}
		}

	}()
	return outChan
}
