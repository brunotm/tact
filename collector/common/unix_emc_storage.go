package common

import (
	"github.com/brunotm/rexon"
	"github.com/brunotm/tact"
	"github.com/brunotm/tact/collector"
)

// NewUnixEMCStorageFn creates a new unix emc storage Fn
func NewUnixEMCStorageFn(inqPath, rex string) tact.GetDataFn {

	storageParser := &rexon.RexLine{
		Rex:    rexon.RexMustCompile(rex),
		Fields: []string{"device", "array_id", "array_device", "device_wwn"},
		Types:  map[string]rexon.ValueType{rexon.KeyTypeAll: rexon.TypeString},
	}

	return func(session *tact.Session) <-chan []byte {
		outChan := make(chan []byte)
		go func() {
			defer close(outChan)
			for _, ct := range []string{" -sym_wwn", " -clar_wwn"} {
				for event := range collector.SSHRex(session, inqPath+ct, storageParser) {
					tact.WrapCtxSend(session.Context(), outChan, event)
				}
			}
		}()
		return outChan
	}
}
