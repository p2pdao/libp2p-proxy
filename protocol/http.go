package protocol

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func (p *ProxyService) httpHandler(bs *BufReaderStream) {
	req, err := http.ReadRequest(bs.Reader)
	if err != nil {
		Log.Error(err)
		writeHTTPError(bs, 400, err)
		bs.CloseWrite()
		return
	}

	isConnectProxy := strings.ToUpper(req.Method) == "CONNECT"
	if !isConnectProxy && !strings.HasPrefix(req.RequestURI, "http://") {
		err = fmt.Errorf("invalid http proxy request: %s, %s, %s", req.Method, req.Host, req.RequestURI)
		writeHTTPError(bs, 400, err)
		bs.CloseWrite()
		return
	}

	if p.isP2PHttp(req.Host) {
		if isConnectProxy {
			fmt.Fprintf(bs, "HTTP/1.1 200 Connection Established\r\n\r\n")
			p.p2phttpHandler(bs, nil)
		} else {
			p.p2phttpHandler(bs, req)
		}
		return
	}

	host := req.Host
	_, port, _ := net.SplitHostPort(req.Host)
	if port == "" {
		host = net.JoinHostPort(req.Host, "80")
	}
	conn, err := net.Dial("tcp", host)
	if err != nil {
		Log.Error(err)
		writeHTTPError(bs, 502, err)
		bs.CloseWrite()
		return
	}

	defer conn.Close()
	if isConnectProxy {
		fmt.Fprintf(bs, "HTTP/1.1 200 Connection Established\r\n\r\n")
	} else {
		go func() {
			req.Header.Set("Connection", req.Header.Get("Proxy-Connection"))
			req.Header.Del("Proxy-Connection")
			err := req.Write(conn)
			if err != nil {
				conn.Close()
			}
			if req.Body != nil {
				req.Body.Close()
			}
		}()
	}

	Log.Infof("http proxying: %s", req.Host)
	if err := tunneling(conn, bs); err != nil && err != io.EOF {
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
