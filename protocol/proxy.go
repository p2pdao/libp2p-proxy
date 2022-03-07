package protocol

import (
	"context"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
)

const (
	ID          = "/p2pdao/libp2p-proxy/1.0.0"
	ServiceName = "p2pdao.libp2p-proxy"
)

var Log = logging.Logger("libp2p-proxy")

type ProxyService struct {
	ctx  context.Context
	host host.Host
	acl  *ACLFilter
}

func NewProxyService(ctx context.Context, h host.Host, acl *ACLFilter) *ProxyService {
	ps := &ProxyService{ctx, h, acl}
	h.SetStreamHandler(ID, ps.Handler)
	return ps
}

func (p *ProxyService) Wait(fn func() error) error {
	<-p.ctx.Done()
	defer p.host.Close()

	if fn != nil {
		if err := fn(); err != nil {
			return err
		}
	}
	return p.ctx.Err()
}

func (p *ProxyService) Handler(s network.Stream) {
	defer s.Close()

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
	b, err := bs.Peek(1)
	if err != nil {
		Log.Errorf("read stream error: %s", err)
		s.Reset()
		return
	}

	if IsSocks5(b[0]) {
		Socks5Handler(bs)
	} else {
		HttpHandler(bs)
	}
}
