package protocol

import (
	"io"
	"time"
)

// Stream
type stream interface {
	io.Reader
	io.Writer
	io.Closer

	SetDeadline(time.Time) error
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

type closeWriter interface {
	CloseWrite() error
}

func tunneling(dst, src stream) error {
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
