package storage

import (
	"github.com/golang/snappy"
)

func SnappyEncode(value []byte) (block []byte) {
	block = make([]byte, snappy.MaxEncodedLen(len(value)))
	return snappy.Encode(block, value)

}

func SnappyDecode(block []byte) (value []byte, err error) {
	sz, err := snappy.DecodedLen(block)
	if err != nil {
		return nil, err
	}

	value = make([]byte, sz)
	value, err = snappy.Decode(value, block)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func CopyBytes(b []byte) (c []byte) {
	c = make([]byte, len(b))
	copy(c, b)
	return c
}
