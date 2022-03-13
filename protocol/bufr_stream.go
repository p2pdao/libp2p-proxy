package protocol

import (
	"bufio"
	"io"
	"time"
)

var _ Stream = (*BufReaderStream)(nil)

type Stream interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error

	SetDeadline(t time.Time) error
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type closeWriter interface {
	CloseWrite() error
}

type closeReader interface {
	CloseRead() error
}

type reseter interface {
	Reset() error
}

type BufReaderStream struct {
	s      Stream
	Reader *bufio.Reader
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
	if s, ok := bs.s.(reseter); ok {
		return s.Reset()
	}
	return bs.s.Close()
}

func (bs *BufReaderStream) CloseWrite() error {
	if s, ok := bs.s.(closeWriter); ok {
		return s.CloseWrite()
	}
	return bs.s.Close()
}

func (bs *BufReaderStream) CloseRead() error {
	if s, ok := bs.s.(closeReader); ok {
		return s.CloseRead()
	}
	return bs.s.Close()
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

func NewBufReaderStream(s Stream) *BufReaderStream {
	return &BufReaderStream{
		s:      s,
		Reader: bufio.NewReader(s),
	}
}

func tunneling(dst, src Stream) error {
	errCh := make(chan error, 2)
	go proxy(dst, src, errCh)
	go proxy(src, dst, errCh)
	// Wait
	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			// return from this function closes target (and conn).
			return err
		}
	}
	return nil
}

func proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	if tcpConn, ok := dst.(closeWriter); ok {
		tcpConn.CloseWrite()
	}
	errCh <- err
}
