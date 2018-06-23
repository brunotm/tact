package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
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
var memoryParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(
			"mem_size_mb",
			rexon.Number,
			rexon.ValueRegex(`memory\s+([-+]?[0-9]*\.?[0-9]+)\s+`)),
		rexon.MustNewValue(
			"mem_used_mb",
			rexon.Number,
			rexon.ValueRegex(`memory\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`)),
		rexon.MustNewValue(
			"mem_free_mb",
			rexon.Number,
			rexon.ValueRegex(`memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`)),
		rexon.MustNewValue(
			"mem_pin_mb",
			rexon.Number,
			rexon.ValueRegex(`memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`)),
		rexon.MustNewValue(
			"mem_virtual_mb",
			rexon.Number,
			rexon.ValueRegex(`memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`)),
		rexon.MustNewValue(
			"mem_avail_mb",
			rexon.Number,
			rexon.ValueRegex(`memory\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)\s+`)),
		rexon.MustNewValue(
			"mem_mode",
			rexon.String,
			rexon.ValueRegex(`memory.*[-+]?[0-9]*\.?[0-9]+\s+(\w+)`)),
		rexon.MustNewValue(
			"swap_size_mb",
			rexon.Number,
			rexon.ValueRegex(`pg\s+space\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"swap_used_mb",
			rexon.Number,
			rexon.ValueRegex(`pg\s+space\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"pin_work_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"pin_pers_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"pin_clnt_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"pin_other_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"used_work_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"used_pers_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`)),
		rexon.MustNewValue(
			"used_clnt_mb",
			rexon.Number,
			rexon.ValueRegex(`pin\s+[-+]?[0-9]*\.?[0-9]+\s+[-+]?[0-9]*\.?[0-9]+\s+([-+]?[0-9]*\.?[0-9]+)`)),
	},
	rexon.StartTag(`^memory\s+\d+`),
)

// System memory collector
func memoryFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, memoryCmd, memoryParser)
}

// User memory collector
var memoryUser = &tact.Collector{
	Name:    "/aix/performance/memory_user",
	GetData: memoryUserFn,
}

// User memory parser
var memoryUserParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue("user", rexon.String),
		rexon.MustNewValue("mem_used_mb", rexon.String),
		rexon.MustNewValue("mem_pin_mb", rexon.String),
		rexon.MustNewValue("swap_used_mb", rexon.String),
		rexon.MustNewValue("mem_virtual_mb", rexon.String),
	},
	rexon.LineRegex(`(\w+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)\s+([-+]?[0-9]*\.?[0-9]+)`),
)

// MemoryUser collector
func memoryUserFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, memoryUserCmd, memoryUserParser)
}
