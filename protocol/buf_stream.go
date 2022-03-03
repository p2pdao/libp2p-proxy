package protocol

import (
	"bufio"

	"github.com/libp2p/go-libp2p-core/network"
)

type BufStream struct {
	network.Stream
	*bufio.Reader
}

func NewBufStream(s network.Stream) *BufStream {
	return &BufStream{
		Stream: s,
		Reader: bufio.NewReader(s),
	}
}
