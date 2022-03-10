package protocol

import (
	"bufio"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
)

var _ network.Stream = (*BufReaderStream)(nil)

type BufReaderStream struct {
	s      network.Stream
	Reader *bufio.Reader
}

func (bs *BufReaderStream) ID() string {
	return bs.s.ID()
}

func (bs *BufReaderStream) Protocol() protocol.ID {
	return bs.s.Protocol()
}
func (bs *BufReaderStream) SetProtocol(id protocol.ID) error {
	return bs.s.SetProtocol(id)
}

func (bs *BufReaderStream) Stat() network.Stats {
	return bs.s.Stat()
}

func (bs *BufReaderStream) Conn() network.Conn {
	return bs.s.Conn()
}

func (bs *BufReaderStream) Scope() network.StreamScope {
	return bs.s.Scope()
}

func (bs *BufReaderStream) Read(p []byte) (int, error) {
	return bs.Reader.Read(p)
}

func (bs *BufReaderStream) Write(p []byte) (n int, err error) {
	return bs.s.Write(p)
}

func (bs *BufReaderStream) Close() error {
	return bs.s.Close()
}

func (bs *BufReaderStream) Reset() error {
	return bs.s.Reset()
}
func (bs *BufReaderStream) CloseWrite() error {
	return bs.s.CloseWrite()
}

func (bs *BufReaderStream) CloseRead() error {
	return bs.s.CloseRead()
}

func (bs *BufReaderStream) SetDeadline(t time.Time) error {
	return bs.s.SetDeadline(t)
}
func (bs *BufReaderStream) SetReadDeadline(t time.Time) error {
	return bs.s.SetReadDeadline(t)
}
func (bs *BufReaderStream) SetWriteDeadline(t time.Time) error {
	return bs.s.SetWriteDeadline(t)
}

func NewBufReaderStream(s network.Stream) *BufReaderStream {
	return &BufReaderStream{
		s:      s,
		Reader: bufio.NewReader(s),
	}
}
