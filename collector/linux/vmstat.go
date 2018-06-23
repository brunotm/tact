package linux

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
	"github.com/brunotm/tact/collector/keys"
	"github.com/brunotm/tact/js"
	"github.com/brunotm/tact/util"
)

const (
	vmstatCmd = "vmstat -s -S M"
)

func init() {
	tact.Registry.Add(vmStat)
}

var vmStat = &tact.Collector{
	Name:    "/linux/performance/vmstat",
	GetData: vmStatFn,
	PostOps: vmStatPostOps,
	EventOps: &tact.EventOps{
		Round: 2,
		Delta: &tact.DeltaOps{
			TTL:  15 * time.Minute,
			Rate: true,
			Blacklist: tact.BuildBlackList(
				keys.MemTotalMB,
				keys.MemUsedMB,
				keys.MemActiveMB,
				keys.MemInactiveMB,
				keys.MemFreeMB,
				keys.MemBufferMB,
				keys.SwapCacheMB,
				keys.SwapTotalMB,
				keys.SwapUsedMB,
				keys.SwapFreeMB),
			RateBlacklist: tact.BuildBlackList(
				"cpu_user_ticks",
				"cpu_usernice_ticks",
				"cpu_system_ticks",
				"cpu_idle_ticks",
				"cpu_wait_ticks"),
		},
	},
}

var vmStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.MemTotalMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+total\s+memory`)),
		rexon.MustNewValue(keys.MemUsedMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+used\s+memory`)),
		rexon.MustNewValue(keys.MemActiveMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+active\s+memory`)),
		rexon.MustNewValue(keys.MemInactiveMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+inactive\s+memory`)),
		rexon.MustNewValue(keys.MemFreeMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+free\s+memory`)),
		rexon.MustNewValue(keys.MemBufferMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+buffer\s+memory`)),
		rexon.MustNewValue(keys.SwapCacheMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+swap\s+cache`)),
		rexon.MustNewValue(keys.SwapTotalMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+total\s+swap`)),
		rexon.MustNewValue(keys.SwapUsedMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+used\s+swap`)),
		rexon.MustNewValue(keys.SwapFreeMB, rexon.Number, rexon.ValueRegex(`(\d+)\s+M?\s+free\s+swap`)),
		rexon.MustNewValue(keys.OSForksAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+forks`)),
		rexon.MustNewValue("cpu_user_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+non-nice\s+user\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_usernice_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+nice\s+user\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_system_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+system\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_idle_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+idle\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_wait_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+IO-wait\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_irq_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+IRQ\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_softirq_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+softirq\s+cpu\s+ticks`)),
		rexon.MustNewValue("cpu_stolen_ticks", rexon.Number, rexon.ValueRegex(`(\d+)\s+stolen\s+cpu\s+ticks`)),
		rexon.MustNewValue(keys.PagesPagedInAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+pages\s+paged\s+in`)),
		rexon.MustNewValue(keys.PagesPagedOutAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+pages\s+paged\s+out`)),
		rexon.MustNewValue(keys.PagesSwappedInAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+pages\s+swapped\s+in`)),
		rexon.MustNewValue(keys.PagesSwappedOutAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+pages\s+swapped\s+out`)),
		rexon.MustNewValue(keys.OSInterruptsAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+interrupts`)),
		rexon.MustNewValue(keys.OSContextSwitchesAvg, rexon.Number, rexon.ValueRegex(`(\d+)\s+CPU\s+context\s+switches`)),
	},
	rexon.StartTag(`\d+\s+M?\s+total\s+memory`),
)

func vmStatFn(ctx *tact.Context) (events <-chan []byte) {
	return ssh.Regex(ctx, vmstatCmd, vmStatParser)
}

func vmStatPostOps(event []byte) (out []byte, err error) {

	userTicks, err := js.GetFloat(event, "cpu_user_ticks")
	userNiceTicks, err := js.GetFloat(event, "cpu_usernice_ticks")
	systemTicks, err := js.GetFloat(event, "cpu_system_ticks")
	idleTicks, err := js.GetFloat(event, "cpu_idle_ticks")
	waitTicks, err := js.GetFloat(event, "cpu_wait_ticks")
	irqTicks, err := js.GetFloat(event, "cpu_irq_ticks")
	sirqTicks, err := js.GetFloat(event, "cpu_softirq_ticks")
	stealTicks, err := js.GetFloat(event, "cpu_stolen_ticks")

	cpuTotal := userTicks + userNiceTicks + systemTicks + idleTicks + waitTicks + irqTicks + sirqTicks + stealTicks
	event, err = js.Set(event, util.Round((userTicks+userNiceTicks)/cpuTotal*100, 2), keys.CPUUserPctAvg)
	event, err = js.Set(event, util.Round((systemTicks+irqTicks+sirqTicks)/cpuTotal*100, 2), keys.CPUSysPctAvg)
	event, err = js.Set(event, util.Round((waitTicks)/cpuTotal*100, 2), keys.CPUWaitPctAvg)
	event, err = js.Set(event, util.Round((stealTicks)/cpuTotal*100, 2), keys.CPUStealWaitPctAvg)

	cpuIdle := util.Round((idleTicks)/cpuTotal*100, 2)
	event, err = js.Set(event, cpuIdle, keys.CPUIdlePctAvg)
	event, err = js.Set(event, util.Round(100-cpuIdle, 2), keys.CPUUsedPctAvg)

	event = js.Delete(event, "cpu_user_ticks")
	event = js.Delete(event, "cpu_usernice_ticks")
	event = js.Delete(event, "cpu_system_ticks")
	event = js.Delete(event, "cpu_idle_ticks")
	event = js.Delete(event, "cpu_wait_ticks")
	event = js.Delete(event, "cpu_irq_ticks")
	event = js.Delete(event, "cpu_softirq_ticks")
	event = js.Delete(event, "cpu_stolen_ticks")

	return event, err
}
