package tact

import (
	"context"
	"encoding/hex"

	blake2b "github.com/minio/blake2b-simd"
)

// WrapCtxSend send a given []byte event in a select with the given context
func WrapCtxSend(ctx context.Context, evtChan chan<- []byte, event []byte) (ok bool) {
	select {
	case <-ctx.Done():
		return false
	case evtChan <- event:
		return true
	}
}

// BuildBlackList function
func BuildBlackList(param ...string) (bl map[string]struct{}) {
	m := make(map[string]struct{})
	var e struct{}
	for _, x := range param {
		m[x] = e
	}
	return m
}

// Hash function
func Hash(v []byte) (h string) {
	b2b := blake2b.New256()
	b2b.Write(v)
	return hex.EncodeToString(b2b.Sum(nil))
}
