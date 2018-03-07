package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	memoryCmd     = "svmon -O unit=MB"
	memoryUserCmd = "svmon -U -O unit=MB"
)

func init() {
	tact.Registry.Add(memory)
	tact.Registry.Add(memoryUser)
}

// System memory collector
var memory = &tact.Collector{
	Name:    "/aix/performance/memory",
	GetData: memoryFn,
}

// System memory parser
var memoryParser = &rexon.RexSet{
	Prep: rexon.RexMustCompile(`[(),!"]`),
	Set: rexon.RexSetMustCompile(map[string]string{
		rexon.KeyStartTag: `^memory\s+\d+`,
		"mem_size_mb":     `^memory\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"mem_used_mb":     `^memory\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"mem_free_mb":     `^memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"mem_pin_mb":      `^memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"mem_virtual_mb":  `^memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"mem_avail_mb": `^memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+
		[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`,
		"mem_mode":     `^memory.*[-+]?[0-9]*\.?[0-9]+\s+(\w+)`,
		"swap_size_mb": `^pg\s+space\s+([-+]?[0-9]*\.?[0-9]+)`,
		"swap_used_mb": `^pg\s+space\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`,
		"pin_work_mb":  `^pin\s+([-+]?[0-9]*\.?[0-9]+)`,
		"pin_pers_mb":  `^pin\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`,
		"pin_clnt_mb":  `^pin\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`,
		"pin_other_mb": `^pin\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`,
		"used_work_mb": `^pin\s+([-+]?[0-9]*\.?[0-9]+)`,
		"used_pers_mb": `^pin\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`,
		"used_clnt_mb": `^pin\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`}),
	Round: 2,
	Types: map[string]rexon.ValueType{
		"mem_mode":       rexon.TypeString,
		rexon.KeyTypeAll: rexon.TypeNumber},
}

// System memory collector
func memoryFn(session *tact.Session) (events <-chan []byte) {
	return collector.SSHRex(session, memoryCmd, memoryParser)
}

// User memory collector
var memoryUser = &tact.Collector{
	Name:    "/aix/performance/memory_user",
	GetData: memoryUserFn,
}

// User memory parser
var memoryUserParser = &rexon.RexLine{
	Rex:    rexon.RexMustCompile(`^(\w+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)`),
	Fields: []string{"user", "mem_used_mb", "mem_pin_mb", "swap_used_mb", "mem_virtual_mb"},
	Round:  2,
	Types: map[string]rexon.ValueType{
		"user":           rexon.TypeString,
		rexon.KeyTypeAll: rexon.TypeNumber},
}

// MemoryUser collector
func memoryUserFn(session *tact.Session) (events <-chan []byte) {
	return collector.SSHRex(session, memoryUserCmd, memoryUserParser)
}
