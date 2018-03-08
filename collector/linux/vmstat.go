package linux

import (
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

const (
	vmstatCmd = "vmstat -sSK"
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
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
		},
		Delta: &tact.DeltaOps{
			TTL:  15 * time.Minute,
			Rate: true,
			Blacklist: tact.BuildBlackList(
				"mem_total_kb",
				"mem_used_kb",
				"mem_active_kb",
				"mem_inactive_kb",
				"mem_free_kb",
				"mem_buffer_kb",
				"swap_cache_kb",
				"swap_total_kb",
				"swap_used_kb",
				"swap_free_kb"),
			RateBlacklist: tact.BuildBlackList(
				"cpu_user_ticks",
				"cpu_usernice_ticks",
				"cpu_system_ticks",
				"cpu_idle_ticks",
				"cpu_wait_ticks"),
		},
	},
}

var vmStatParser = &rexon.RexSet{
	Set: rexon.RexSetMustCompile(map[string]string{
		rexon.KeyStartTag:          `^\d+\s+K?\s+total\s+memory`,
		"mem_total_kb":             `^(\d+)\s+K?\s+total\s+memory`,
		"mem_used_kb":              `^(\d+)\s+K?\s+used\s+memory`,
		"mem_active_kb":            `^(\d+)\s+K?\s+active\s+memory`,
		"mem_inactive_kb":          `^(\d+)\s+K?\s+inactive\s+memory`,
		"mem_free_kb":              `^(\d+)\s+K?\s+free\s+memory`,
		"mem_buffer_kb":            `^(\d+)\s+K?\s+buffer\s+memory`,
		"swap_cache_kb":            `^(\d+)\s+K?\s+swap\s+cache`,
		"swap_total_kb":            `^(\d+)\s+K?\s+total\s+swap`,
		"swap_used_kb":             `^(\d+)\s+K?\s+used\s+swap`,
		"swap_free_kb":             `^(\d+)\s+K?\s+free\s+swap`,
		"avg_forks":                `^(\d+)\s+forks`,
		"cpu_user_ticks":           `^(\d+)\s+non-nice\s+user\s+cpu\s+ticks`,
		"cpu_usernice_ticks":       `^(\d+)\s+nice\s+user\s+cpu\s+ticks`,
		"cpu_system_ticks":         `^(\d+)\s+system\s+cpu\s+ticks`,
		"cpu_idle_ticks":           `^(\d+)\s+idle\s+cpu\s+ticks`,
		"cpu_wait_ticks":           `^(\d+)\s+IO-wait\s+cpu\s+ticks`,
		"cpu_irq_ticks":            `^(\d+)\s+IRQ\s+cpu\s+ticks`,
		"cpu_softirq_ticks":        `^(\d+)\s+softirq\s+cpu\s+ticks`,
		"cpu_stolen_ticks":         `^(\d+)\s+stolen\s+cpu\s+ticks`,
		"avg_pages_paged_in":       `^(\d+)\s+pages\s+paged\s+in`,
		"avg_pages_paged_out":      `^(\d+)\s+pages\s+paged\s+out`,
		"avg_pages_swapped_in":     `^(\d+)\s+pages\s+swapped\s+in`,
		"avg_pages_swapped_out":    `^(\d+)\s+pages\s+swapped\s+out`,
		"avg_interrupts":           `^(\d+)\s+interrupts`,
		"avg_cpu_context_switches": `^(\d+)\s+CPU\s+context\s+switches`,
	}),
}

func vmStatFn(session *tact.Session) (events <-chan []byte) {
	return collector.SSHRex(session, vmstatCmd, vmStatParser)
}

func vmStatPostOps(event []byte) (out []byte, err error) {

	userTicks, err := rexon.JSONGetFloat(event, "cpu_user_ticks")
	userNiceTicks, err := rexon.JSONGetFloat(event, "cpu_usernice_ticks")
	systemTicks, err := rexon.JSONGetFloat(event, "cpu_system_ticks")
	idleTicks, err := rexon.JSONGetFloat(event, "cpu_idle_ticks")
	waitTicks, err := rexon.JSONGetFloat(event, "cpu_wait_ticks")
	irqTicks, err := rexon.JSONGetFloat(event, "cpu_irq_ticks")
	sirqTicks, err := rexon.JSONGetFloat(event, "cpu_softirq_ticks")
	stealTicks, err := rexon.JSONGetFloat(event, "cpu_stolen_ticks")

	cpuTotal := userTicks + userNiceTicks + systemTicks + idleTicks + waitTicks + irqTicks + sirqTicks + stealTicks
	event, err = rexon.JSONSet(event, rexon.Round((userTicks+userNiceTicks)/cpuTotal*100, 2), "avg_cpu_user")
	event, err = rexon.JSONSet(event, rexon.Round((systemTicks+irqTicks+sirqTicks)/cpuTotal*100, 2), "avg_cpu_system")
	event, err = rexon.JSONSet(event, rexon.Round((waitTicks)/cpuTotal*100, 2), "avg_cpu_iowait")
	event, err = rexon.JSONSet(event, rexon.Round((stealTicks)/cpuTotal*100, 2), "avg_cpu_stealwait")

	cpuIdle := rexon.Round((idleTicks)/cpuTotal*100, 2)
	event, err = rexon.JSONSet(event, cpuIdle, "avg_cpu_idle")
	event, err = rexon.JSONSet(event, rexon.Round(100-cpuIdle, 2), "avg_cpu_utilization")

	event = rexon.JSONDelete(event, "cpu_user_ticks")
	event = rexon.JSONDelete(event, "cpu_usernice_ticks")
	event = rexon.JSONDelete(event, "cpu_system_ticks")
	event = rexon.JSONDelete(event, "cpu_idle_ticks")
	event = rexon.JSONDelete(event, "cpu_wait_ticks")
	event = rexon.JSONDelete(event, "cpu_irq_ticks")
	event = rexon.JSONDelete(event, "cpu_softirq_ticks")
	event = rexon.JSONDelete(event, "cpu_stolen_ticks")

	return event, err
}
