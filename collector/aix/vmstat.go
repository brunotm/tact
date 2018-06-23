package aix

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/collector/keys"
)

const (
	vmStatCmd = "/usr/bin/vmstat -w 60 1"
)

func init() {
	tact.Registry.Add(vmStat)
}

var vmStat = &tact.Collector{
	Name:    "/aix/performance/vmstat",
	GetData: vmStatFn,
}

var vmStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.OSThreadsRunnableAvg, rexon.Number),
		rexon.MustNewValue(keys.OSThreadsBlockedAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesActiveAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesFreeAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesReclaimedAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesPagedInAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesPagedOutAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesFreedAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesScannedAvg, rexon.Number),
		rexon.MustNewValue(keys.PagesCyclesAvg, rexon.Number),
		rexon.MustNewValue(keys.OSInterruptsAvg, rexon.Number),
		rexon.MustNewValue(keys.OSSyscallsAvg, rexon.Number),
		rexon.MustNewValue(keys.OSContextSwitchesAvg, rexon.Number),
		rexon.MustNewValue(keys.CPUUserPctAvg, rexon.Number),
		rexon.MustNewValue(keys.CPUSysPctAvg, rexon.Number),
		rexon.MustNewValue(keys.CPUIdlePctAvg, rexon.Number),
		rexon.MustNewValue(keys.CPUWaitPctAvg, rexon.Number),
		rexon.MustNewValue(keys.CPUPhysicalConsumedAvg, rexon.Number),
		rexon.MustNewValue(keys.CPUEntitledCapacityPctAvg, rexon.Number),
	},
	rexon.LineRegex(`([-+]?\d*\.?\d+)`), rexon.FindAll(),
)

// VMStat collector
func vmStatFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, vmStatCmd, vmStatParser)
}
