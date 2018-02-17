package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	vmStatCmd = "/usr/bin/vmstat -IWw 60 1"
)

// init register this collector with the dispatcher
func init() {
	tact.Registry.Add(vmStat)
}

var vmStat = &tact.Collector{
	Name:    "/aix/performance/vmstat",
	GetData: vmStatFn,
}

var vmStatParser = &rexon.RexLine{
	FindAll: true,
	Rex:     rexon.RexMustCompile(`([-+]?\d*\.?\d+)`),
	Fields: []string{"kthr_r", "kthr_b", "kthr_p", "kthr_w", "mem_avm", "mem_fre",
		"page_fi", "page_fo", "page_pi", "page_po", "page_fr", "page_sr",
		"faults_in", "faults_sy", "faults_cs", "cpu_us", "cpu_sy", "cpu_id",
		"cpu_wa", "cpu_pc", "cpu_ec"},
	Round: 2,
	Types: map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeNumber},
}

// VMStat collector
func vmStatFn(session *tact.Session) <-chan []byte {
	return collector.SSHRex(session, vmStatCmd, vmStatParser)
}
