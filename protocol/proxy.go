package protocol

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
	gostream "github.com/libp2p/go-libp2p-gostream"
)

const (
	P2PHttpID   protocol.ID = "/http"
	ID          protocol.ID = "/p2pdao/libp2p-proxy/1.0.0"
	ServiceName string      = "p2pdao.libp2p-proxy"
)

var Log = logging.Logger("libp2p-proxy")

type ProxyService struct {
	ctx     context.Context
	host    host.Host
	acl     *ACLFilter
	http    *http.Server
	p2pHost string
}

func NewProxyService(ctx context.Context, h host.Host, acl *ACLFilter, p2pHost string) *ProxyService {
	ps := &ProxyService{ctx, h, acl, nil, p2pHost}
	h.SetStreamHandler(ID, ps.Handler)
	return ps
}

// Close terminates this listener. It will no longer handle any
// incoming streams
func (p *ProxyService) Close() error {
	if s := p.http; s != nil {
		c, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		p.http = nil
		s.Shutdown(c)
	}
	return p.host.Close()
}

func (p *ProxyService) Wait(fn func() error) error {
	<-p.ctx.Done()
	defer p.Close()

	if fn != nil {
		if err := fn(); err != nil {
			return err
		}
	}
	return p.ctx.Err()
}

func (p *ProxyService) Handler(s network.Stream) {
	// defer s.Close()

	if p.acl != nil && !p.acl.Allow(s.Conn().RemotePeer(), s.Conn().RemoteMultiaddr()) {
		Log.Infof("refusing proxy for %s; permission denied", s.Conn().RemotePeer())
		s.Reset()
		return
	}

	if err := s.Scope().SetService(ServiceName); err != nil {
		Log.Errorf("error attaching stream to service: %s", err)
		s.Reset()
		return
	}

	bs := NewBufReaderStream(s)
	b, err := bs.Reader.Peek(1)
	if err != nil {
		Log.Errorf("read stream error: %s", err)
		s.Reset()
		return
	}

	if IsSocks5(b[0]) {
		p.socks5Handler(bs)
	} else {
		p.httpHandler(bs)
	}
}

func (p *ProxyService) ServeHTTP(handler http.Handler, s *http.Server) error {
	if p.http != nil {
		return fmt.Errorf("http.Server exists")
	}
	if handler == nil {
		return fmt.Errorf("http handler is nil")
	}
	l, err := gostream.Listen(p.host, P2PHttpID)
	if err != nil {
		return err
	}

	if s == nil {
		s = new(http.Server)
		s.ReadHeaderTimeout = 20 * time.Second
		s.ReadTimeout = 60 * time.Second
		s.WriteTimeout = 120 * time.Second
		s.IdleTimeout = 90 * time.Second
	}
	s.Handler = handler
	p.http = s

	go p.Wait(nil)
	return s.Serve(l)
}

func (p *ProxyService) isP2PHttp(host string) bool {
	return strings.HasPrefix(host, p.p2pHost)
}
