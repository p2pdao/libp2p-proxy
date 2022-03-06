package protocol

import (
	"bufio"

	"github.com/libp2p/go-libp2p-core/network"
)

type BufReaderStream struct {
	network.Stream
	*bufio.Reader
}

func (bs *BufReaderStream) Read(p []byte) (int, error) {
	return bs.Reader.Read(p)
}

func NewBufReaderStream(s network.Stream) *BufReaderStream {
	return &BufReaderStream{
		Stream: s,
		Reader: bufio.NewReader(s),
	}
}
