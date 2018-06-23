package linux

import (
	"fmt"
	"time"

	"github.com/brunotm/tact/collector/keys"

	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/sftp"
	"github.com/brunotm/tact/js"
)

const (
	fileName   = "messages"
	timeLayout = "Jan 2 15:04:05 2006"
)

func init() {
	tact.Registry.Add(logMessages)
}

var logMessages = &tact.Collector{
	Name:    "/linux/log/messages",
	GetData: logMessagesFn,
	PostOps: logMessagesPostOps,
}

var logMessagesParser = rexon.MustNewParser(
	[]*rexon.Value{
		rexon.MustNewValue(keys.Time, rexon.String),
		rexon.MustNewValue("resource", rexon.String),
		rexon.MustNewValue("pid", rexon.Number),
		rexon.MustNewValue("message", rexon.String),
	},
)

func logMessagesFn(ctx *tact.Context) (events <-chan []byte) {
	return sftp.Regex(ctx, fileName, logMessagesParser)
}

func logMessagesPostOps(event []byte) (out []byte, err error) {
	ts, _ := js.GetUnsafeString(event, keys.Time)
	timestamp, err := time.Parse(timeLayout, fmt.Sprintf("%s %d", ts, time.Now().Year()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %s, error: %s", ts, err.Error())
	}
	event, err = js.Set(event, timestamp, keys.Time)
	return event, err
}
