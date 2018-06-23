package aix

import (
	"bytes"
	"time"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
)

const (
	fcListCmd = `lsdev -l fcs* -F name`
	fcStatCmd = "fcstat "
)

func init() {
	tact.Registry.Add(fcStat)
}

var fcStat = &tact.Collector{
	Name:    "/aix/performance/fcstat",
	GetData: fcStatFn,
	EventOps: &tact.EventOps{
		Round: 2,
		Delta: &tact.DeltaOps{
			TTL:      time.Hour * 3,
			KeyField: "device",
			Rate:     true,
			Blacklist: tact.BuildBlackList(
				"device",
				"serial_number",
				"wwpn",
				"speed_sup_gbit",
				"speed_run_gbit",
				"fcid",
				"last_reset_sec"),
			RateBlacklist: tact.BuildBlackList(
				"num_lip_count",
				"num_nos_count",
				"num_frames_error",
				"num_frames_dumped",
				"num_link_fail",
				"num_sync_loss",
				"num_prim_seq_error",
				"num_invalid_tx_word",
				"num_invalid_crc",
				"num_fc_no_dma_res",
				"num_fc_no_adapt_element",
				"num_fc_no_cmd_res",
				"num_fc_cntrl_req"),
		},
	},
}

var fcStatParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue("device", rexon.String, rexon.ValueRegex(`FIBRE\s+CHANNEL\s+STATISTICS\s+REPORT:\s+(\w+)`)),
		rexon.MustNewValue("serial_number", rexon.String, rexon.ValueRegex(`^Serial\s+Number:\s+(\w+)`)),
		rexon.MustNewValue("wwpn", rexon.String, rexon.ValueRegex(`World\s+Wide\s+Port\s+Name:\s+0x(\w+)`)),
		rexon.MustNewValue("speed_sup_gbit", rexon.Number, rexon.ValueRegex(`Port\s+Speed\s+\(supported\):\s+(\w+)\s+GBIT`)),
		rexon.MustNewValue("speed_run_gbit", rexon.Number, rexon.ValueRegex(`Port\s+Speed\s+\(running\):\s+(\w+)\s+GBIT`)),
		rexon.MustNewValue("fcid", rexon.String, rexon.ValueRegex(`Port\s+FC\s+ID:\s+(\w+)`)),
		rexon.MustNewValue("last_reset_seconds", rexon.Number, rexon.ValueRegex(`Seconds\s+Since\s+Last\s+Reset:\s+(\d+)`)),
		rexon.MustNewValue("avg_tx_frames", rexon.Number, rexon.ValueRegex(`Frames:\s+(\d+)\s+\d+`)),
		rexon.MustNewValue("avg_rx_frames", rexon.Number, rexon.ValueRegex(`Frames:\s+\d+\s+(\d+)`)),
		rexon.MustNewValue("avg_tx_words", rexon.Number, rexon.ValueRegex(`Words:\s+(\d+)\s+\d+`)),
		rexon.MustNewValue("avg_rx_words", rexon.Number, rexon.ValueRegex(`Words:\s+\d+\s+(\d+)`)),
		rexon.MustNewValue("num_lip_count", rexon.Number, rexon.ValueRegex(`LIP\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_nos_count", rexon.Number, rexon.ValueRegex(`NOS\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_frames_error", rexon.Number, rexon.ValueRegex(`Error\s+Frames:\s+(\d+)`)),
		rexon.MustNewValue("num_frames_dumped", rexon.Number, rexon.ValueRegex(`Dumped\s+Frames:\s+(\d+)`)),
		rexon.MustNewValue("num_link_fail", rexon.Number, rexon.ValueRegex(`Link\s+Failure\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_sync_loss", rexon.Number, rexon.ValueRegex(`Loss\s+of\s+Sync\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_signal_loss", rexon.Number, rexon.ValueRegex(`Loss\s+of\s+Signal:\s+(\d+)`)),
		rexon.MustNewValue("num_prim_seq_error", rexon.Number, rexon.ValueRegex(`Primitive\s+Seq\s+Protocol\s+Error\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_invalid_tx_word", rexon.Number, rexon.ValueRegex(`Invalid\s+Tx\s+Word\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_invalid_crc", rexon.Number, rexon.ValueRegex(`Invalid\s+CRC\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_fc_no_dma_res", rexon.Number, rexon.ValueRegex(`No\s+DMA\s+Resource\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_fc_no_adapt_element", rexon.Number, rexon.ValueRegex(`^No\s+Adapter\s+Elements\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("num_fc_no_cmd_res", rexon.Number, rexon.ValueRegex(`No\s+Command\s+Resource\s+Count:\s+(\d+)`)),
		rexon.MustNewValue("avg_fc_read_req", rexon.Number, rexon.ValueRegex(`Input\s+Requests:\s+(\d+)`)),
		rexon.MustNewValue("avg_fc_write_req", rexon.Number, rexon.ValueRegex(`Output\s+Requests:\s+(\d+)`)),
		rexon.MustNewValue("num_fc_cntrl_req", rexon.Number, rexon.ValueRegex(`Control\s+Requests:\s+(\d+)`)),
		rexon.MustNewValue("avg_fc_megabytes_rx", rexon.DigitalUnit, rexon.ValueRegex(`Input\s+Bytes:\s+(\d+)`),
			rexon.ToFormat("mb")),
		rexon.MustNewValue("avg_fc_megabytes_tx", rexon.DigitalUnit, rexon.ValueRegex(`Output\s+Bytes:\s+(\d+)`),
			rexon.ToFormat("mb")),
	},
	rexon.StartTag(`FIBRE\s+CHANNEL\s+STATISTICS\s+REPORT:\s+(\w+)`),
	rexon.SkipTag(`IP\s+over\s+FC\s+Traffic\s+Statistics`),
	rexon.ContinueTag(`FC\s+SCSI\s+Traffic\s+Statistics`),
)

// FCStat collector
func fcStatFn(ctx *tact.Context) (events <-chan []byte) {
	outCh := make(chan []byte)

	go func() {
		defer close(outCh)

		client, err := ssh.Client(ctx)
		if err != nil {
			ctx.LogError("getting ssh client error: %s", fcListCmd, err)
			return
		}
		defer client.Close()

		fcs, err := client.CombinedOutput(fcListCmd, nil)
		if err != nil {
			ctx.LogError("executing ssh comand: %s, error: %s", fcListCmd, err)
			return
		}

		for _, fc := range bytes.Split(fcs, []byte("\n")) {
			if len(fc) == 0 {
				continue
			}

			fcstat, err := client.CombinedOutput(fcStatCmd+string(fc), nil)
			if err != nil {
				ctx.LogError("executing ssh comand: %s, error: %s", fcStatCmd+string(fc), err)
				continue
			}

			for result := range fcStatParser.ParseBytes(ctx.Context(), fcstat) {
				for e := range result.Errors {
					ctx.LogError(result.Errors[e].Error())
				}

				if !tact.WrapCtxSend(ctx.Context(), outCh, result.Data) {
					ctx.LogError("timeout sending event upstream")
					return
				}
			}
		}
	}()

	return outCh
}
