package protocol

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func HttpHandler(bs *BufReaderStream) {
	req, err := http.ReadRequest(bs.Reader)
	if err != nil {
		Log.Error(err)
		writeHTTPError(bs, 400, err)
		bs.CloseWrite()
		return
	}

	if strings.ToUpper(req.Method) != "CONNECT" {
		err = fmt.Errorf("invalid request method: %s", req.Method)
		writeHTTPError(bs, 400, err)
		bs.CloseWrite()
		return
	}

	conn, err := net.Dial("tcp", req.Host)
	if err != nil {
		Log.Error(err)
		writeHTTPError(bs, 500, err)
		bs.CloseWrite()
		return
	}

	defer conn.Close()
	fmt.Fprintf(bs, "HTTP/1.1 200 Connection Established\r\n\r\n")

	Log.Infof("http proxying: %s", req.Host)
	if err := tunneling(bs, conn); err != nil {
		Log.Error(err)
	}
}

func writeHTTPError(w io.Writer, code int, err error) {
	fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", code, http.StatusText(code))
	fmt.Fprintf(w, "Server: %s\r\n", ServiceName)
	fmt.Fprintf(w, "Date: %s\r\n", time.Now().Format(http.TimeFormat))
	fmt.Fprintf(w, "Content-Type: text/plain; charset=utf-8\r\n")
	msg := err.Error()
	fmt.Fprintf(w, "Content-Length: %d\r\n", len(msg))
	fmt.Fprintf(w, "Connection: close\r\n\r\n")
	fmt.Fprintf(w, "%s\n", msg)
}
