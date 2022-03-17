package protocol

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
)

func (p *ProxyService) p2phttpHandler(bs *BufReaderStream, req *http.Request) {
	var err error
	for {
		bs.SetReadDeadline(time.Now().Add(time.Second * 10))
		if req == nil {
			req, err = http.ReadRequest(bs.Reader)
		}

		if err != nil {
			if err == io.EOF {
				return
			}

			Log.Error(err)
			writeHTTPError(bs, 400, err)
			bs.Reset()
			return
		}

		pp, err := parsePath(req.URL.Path)
		if err != nil {
			err = fmt.Errorf("failed to parse request: %v", err)
			Log.Error(err)
			writeHTTPError(bs, 400, err)
			bs.Reset()
			return
		}

		req.Host = pp.target.ID.String() // Let URL's Host take precedence.
		req.URL.Path = pp.httpPath
		req.Close = true
		if len(pp.target.Addrs) > 0 {
			p.host.Peerstore().AddAddrs(pp.target.ID, pp.target.Addrs, peerstore.TempAddrTTL)
		}
		s, err := p.host.NewStream(req.Context(), pp.target.ID, pp.protocol)
		if err != nil {
			if req.Body != nil {
				req.Body.Close()
			}
			err = fmt.Errorf("dial remote error: %v", err)
			Log.Error(err)
			writeHTTPError(bs, 500, err)
			bs.Reset()
			return
		}

		Log.Infof("p2p proxying: %s", pp.target.ID.String())
		// Write the request while reading the response
		go func() {
			err := req.Write(s)
			if err != nil {
				s.Close()
			}
			if req.Body != nil {
				req.Body.Close()
			}
		}()

		_, err = io.Copy(bs, s)
		s.Close()
		req = nil
		if shouldLogError(err) {
			Log.Warn(err)
		}
	}
}

type proxyPath struct {
	target   *peer.AddrInfo
	protocol protocol.ID
	httpPath string // path to send to the proxy-host
}

// from the url path parse the peer.AddrInfo, protocol and http path
// /p2p/$peer_id/http/$http_path
// or
// /p2p/$peer_id/x/$protocol/http/$http_path
// or
// /ip4/127.0.0.1/tcp/1234/p2p/$peer_id/http/$http_path
func parsePath(path string) (*proxyPath, error) {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "//") {
		return nil, fmt.Errorf("invalid p2p request path: %s", strconv.Quote(path))
	}
	if strings.HasSuffix(path, "/http") && !strings.Contains(path, "/http/") {
		path += "/"
	}

	pp := &proxyPath{protocol: P2PHttpID, httpPath: "/"}
	ss := strings.SplitN(path, "/http/", 2)
	if len(ss) > 1 {
		pp.httpPath += ss[1]
	}

	ss = strings.SplitN(ss[0], "/x/", 2)
	if len(ss) > 1 {
		pp.protocol = protocol.ID("/x/" + ss[1] + "/http")
	}
	ss = strings.SplitN(ss[0], "/p2p/", 2)
	if len(ss) < 2 {
		return nil, fmt.Errorf("invalid request path %s, no \"/p2p/\"", strconv.Quote(path))
	}

	fullAddr := ss[0] + "/p2p/" + ss[1]
	target, err := peer.AddrInfoFromString(fullAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid request peer %s", fullAddr)
	}
	pp.target = target
	return pp, nil
}
