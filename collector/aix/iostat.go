package aix

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	ioStatCmd = "/usr/bin/iostat -DRVl 60 1"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(ioStat)
}

var ioStat = &tact.Collector{
	Name:    "/aix/performance/iostat",
	GetData: ioStatFn,
	EventOps: &tact.EventOps{
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
			"device":         rexon.TypeString,
		},
	},
	Joins: []*tact.Join{
		&tact.Join{
			TTL:           3 * time.Hour,
			Name:          "/aix/config/storage",
			JoinFields:    []string{"device"},
			JoinOnFields:  []string{"device"},
			IncludeFields: []string{"array_id", "array_device", "device_wwn"},
		},
		&tact.Join{
			TTL:           3 * time.Hour,
			Name:          "/aix/config/lspv",
			JoinFields:    []string{"device"},
			JoinOnFields:  []string{"device"},
			IncludeFields: []string{"vg_name", "vg_mode", "pvid"},
		},
	},
}

var ioStatParser = &rexon.RexLine{
	FindAll: true,
	Round:   2,
	Rex:     rexon.RexMustCompile(`(hdisk\d+|[-+]?[0-9]*\.?[0-9]+\w?)`),
	Fields: []string{"device", "tm_act", "kbps", "iops", "kbread", "kbwrtn", "rps",
		"r_avgserv", "r_minserv", "r_maxserv", "r_timeouts", "r_fail", "wps",
		"w_avgserv", "w_minserv", "w_maxserv", "w_timeouts", "w_fail", "avgqt",
		"minqt", "maxqt", "avgwqsz", "avgsqsz", "sqfull"},
}

// iostat collector
func ioStatFn(session tact.Session) <-chan []byte {
	outChan := make(chan []byte)
	inChan := collector.SSHRex(session, ioStatCmd, ioStatParser)

	// Lauch a goroutine to intercept the events and convert
	// the value for keys below from a size string notation to the specified unit
	go func() {
		defer close(outChan)
		for event := range inChan {
			// Parse the human readable values for the given keys
			// Cannot do this as a PostOps because of units parsing
			for _, key := range []string{"kbps", "kbread", "kbwrtn"} {
				if value, _, err := rexon.JSONGet(event, key); err == nil {
					parsed, err := rexon.ParseSize(value, rexon.KBytes)
					if err != nil {
						session.LogErr("parsing value for: %s, error: %s", key, err.Error())
						event = rexon.JSONDelete(event, key)
						continue
					}
					event, _ = rexon.JSONSet(event, parsed, key)

				} else {
					session.LogErr("error getting value for %s, error: %s", key, err.Error())
					event = rexon.JSONDelete(event, key)
					continue
				}
			}
			if !tact.WrapCtxSend(session.Context(), outChan, event) {
				session.LogErr("timeout sending event upstream")
			}
		}
	}()

	return outChan
}
