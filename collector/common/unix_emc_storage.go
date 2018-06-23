package common

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector/client/ssh"
)

// NewUnixEMCStorageFn creates a new unix emc storage Fn
func NewUnixEMCStorageFn(inqPath, expr string) tact.GetDataFn {

	storageParser := rexon.MustNewParser(
		[]*rexon.Value{
			rexon.MustNewValue("device", rexon.String),
			rexon.MustNewValue("array_id", rexon.String),
			rexon.MustNewValue("array_device", rexon.String),
			rexon.MustNewValue("device_wwn", rexon.String),
		},
		rexon.LineRegex(expr),
	)

	return func(ctx *tact.Context) (events <-chan []byte) {
		outCh := make(chan []byte)
		go func() {
			defer close(outCh)
			for _, ct := range []string{" -sym_wwn", " -clar_wwn"} {
				for event := range ssh.Regex(ctx, inqPath+ct, storageParser) {
					tact.WrapCtxSend(ctx.Context(), outCh, event)
				}
			}
		}()
		return outCh
	}
}
