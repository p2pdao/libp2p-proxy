package protocol

import (
	"fmt"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
)

var log = logging.Logger("libp2p-proxy")
var Log = log

const (
	ID          = "/p2pdao/libp2p-proxy/1.0.0"
	ServiceName = "p2pdao.libp2p-proxy"
)

type ProxyService struct {
	host host.Host
	acl  *ACLFilter
}

func NewProxyService(h host.Host, acl *ACLFilter) *ProxyService {
	ps := &ProxyService{h, acl}
	h.SetStreamHandler(ID, ps.Handler)
	return ps
}

func (p *ProxyService) Handler(s network.Stream) {
	if p.acl != nil && !p.acl.Allow(s.Conn().RemotePeer(), s.Conn().RemoteMultiaddr()) {
		log.Infof("refusing proxy for %s; permission denied", s.Conn().RemotePeer())
		s.Reset()
		return
	}

	if err := s.Scope().SetService(ServiceName); err != nil {
		log.Errorf("error attaching stream to service: %s", err)
		s.Reset()
		return
	}

	bs := NewBufStream(s)
	b, err := bs.Peek(1)
	if err != nil {
		log.Errorf("read stream error: %s", err)
		s.Reset()
		return
	}

	if IsSocks5(b[0]) {
		fmt.Println("Socks5 version")
		Socks5Handler(bs)
	} else {
		fmt.Println("Http version")
		HttpHandler(bs)
	}
}
