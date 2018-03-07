package aix

import (
	"bytes"

	"github.com/brunotm/rexon"
	"github.com/brunotm/sshmgr"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
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
		FieldTypes: map[string]rexon.ValueType{
			rexon.KeyTypeAll: rexon.TypeNumber,
			"device":         rexon.TypeString,
			"serial_number":  rexon.TypeString,
			"wwpn":           rexon.TypeString,
			"fcid":           rexon.TypeString,
		},
		Delta: &tact.DeltaOps{
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

var fcStatParser = &rexon.RexSet{
	Set: rexon.RexSetMustCompile(map[string]string{
		rexon.KeyStartTag:         `^FIBRE\s+CHANNEL\s+STATISTICS\s+REPORT:\s+(\w+)`,
		rexon.KeySkipTag:          `^IP\s+over\s+FC\s+Traffic\s+Statistics`,
		rexon.KeyContinueTag:      `^FC\s+SCSI\s+Traffic\s+Statistics`,
		"device":                  `^FIBRE\s+CHANNEL\s+STATISTICS\s+REPORT:\s+(\w+)`,
		"serial_number":           `^Serial\s+Number:\s+(\w+)`,
		"wwpn":                    `World\s+Wide\s+Port\s+Name:\s+0x(\w+)`,
		"speed_sup_gbit":          `^Port\s+Speed\s+\(supported\):\s+(\w+)\s+GBIT`,
		"speed_run_gbit":          `^Port\s+Speed\s+\(running\):\s+(\w+)\s+GBIT`,
		"fcid":                    `^Port\s+FC\s+ID:\s+(\w+)`,
		"last_reset_sec":          `^Seconds\s+Since\s+Last\s+Reset:\s+(\d+)`,
		"avg_tx_frames":           `^Frames:\s+(\d+)\s+\d+`,
		"avg_rx_frames":           `^Frames:\s+\d+\s+(\d+)`,
		"avg_tx_words":            `^Words:\s+(\d+)\s+\d+`,
		"avg_rx_words":            `^Words:\s+\d+\s+(\d+)`,
		"num_lip_count":           `^LIP\s+Count:\s+(\d+)`,
		"num_nos_count":           `^NOS\s+Count:\s+(\d+)`,
		"num_frames_error":        `^Error\s+Frames:\s+(\d+)`,
		"num_frames_dumped":       `^Dumped\s+Frames:\s+(\d+)`,
		"num_link_fail":           `^Link\s+Failure\s+Count:\s+(\d+)`,
		"num_sync_loss":           `^Loss\s+of\s+Sync\s+Count:\s+(\d+)`,
		"num_signal_loss":         `^Loss\s+of\s+Signal:\s+(\d+)`,
		"num_prim_seq_error":      `^Primitive\s+Seq\s+Protocol\s+Error\s+Count:\s+(\d+)`,
		"num_invalid_tx_word":     `^Invalid\s+Tx\s+Word\s+Count:\s+(\d+)`,
		"num_invalid_crc":         `^Invalid\s+CRC\s+Count:\s+(\d+)`,
		"num_fc_no_dma_res":       `^No\s+DMA\s+Resource\s+Count:\s+(\d+)`,
		"num_fc_no_adapt_element": `^No\s+Adapter\s+Elements\s+Count:\s+(\d+)`,
		"num_fc_no_cmd_res":       `^No\s+Command\s+Resource\s+Count:\s+(\d+)`,
		"avg_fc_read_req":         `^Input\s+Requests:\s+(\d+)`,
		"avg_fc_write_req":        `^Output\s+Requests:\s+(\d+)`,
		"num_fc_cntrl_req":        `^Control\s+Requests:\s+(\d+)`,
		"avg_fc_bytes_rx":         `^Input\s+Bytes:\s+(\d+)`,
		"avg_fc_bytes_tx":         `^Output\s+Bytes:\s+(\d+)`,
	}),
}

// FCStat collector
func fcStatFn(session *tact.Session) (events <-chan []byte) {
	outCh := make(chan []byte)

	go func() {
		defer close(outCh)

		sshSession, err := sshmgr.Manager.GetSSHSession(collector.NewSSHNodeConfig(session.Node()))
		if err != nil {
			session.LogErr("error getting ssh session: %s", err)
			return
		}
		defer sshSession.Close()

		fcs, err := sshSession.CombinedOutput(fcListCmd)
		if err != nil {
			session.LogErr("executing ssh comand: %s, error: %s", fcListCmd, err.Error())
			return
		}

		for _, fc := range bytes.Split(fcs, []byte("\n")) {
			if len(fc) == 0 {
				continue
			}

			sess, err := sshmgr.Manager.GetSSHSession(collector.NewSSHNodeConfig(session.Node()))
			if err != nil {
				session.LogErr("error getting ssh session: %s", err)
				return
			}
			defer sess.Close()

			fcstat, err := sess.CombinedOutput(fcStatCmd + string(fc))
			if err != nil {
				session.LogErr("executing ssh comand: %s, error: %s", fcStatCmd+string(fc), err)
				continue
			}

			for result := range fcStatParser.ParseBytes(session.Context(), fcstat) {
				for e := range result.Errors {
					session.LogErr(result.Errors[e].Error())
				}
				if !tact.WrapCtxSend(session.Context(), outCh, result.Data) {
					session.LogErr("timeout sending event upstream")
					return
				}
			}
		}
	}()

	return outCh
}
