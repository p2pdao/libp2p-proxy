package protocol

import (
	"fmt"
	"net"

	"github.com/libp2p/go-libp2p/core/connmgr"
	"github.com/libp2p/go-libp2p/core/control"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"

	"github.com/p2pdao/libp2p-proxy/config"
)

var _ connmgr.ConnectionGater = (*ACLFilter)(nil)

type ACLFilter struct {
	allowPeers   map[peer.ID]struct{}
	allowSubnets []*net.IPNet
}

func NewACL(cfg config.ACLConfig) (*ACLFilter, error) {
	acl := &ACLFilter{}

	if len(cfg.AllowPeers) > 0 {
		acl.allowPeers = make(map[peer.ID]struct{})
		for _, s := range cfg.AllowPeers {
			p, err := peer.Decode(s)
			if err != nil {
				return nil, fmt.Errorf("error parsing peer ID: %w", err)
			}

			acl.allowPeers[p] = struct{}{}
		}
	}

	if len(cfg.AllowSubnets) > 0 {
		acl.allowSubnets = make([]*net.IPNet, 0, len(cfg.AllowSubnets))
		for _, s := range cfg.AllowSubnets {
			_, ipnet, err := net.ParseCIDR(s)
			if err != nil {
				return nil, fmt.Errorf("error parsing subnet: %w", err)
			}
			acl.allowSubnets = append(acl.allowSubnets, ipnet)
		}
	}

	return acl, nil
}

func (a *ACLFilter) Allow(p peer.ID, addr ma.Multiaddr) bool {
	if len(a.allowPeers) > 0 {
		_, ok := a.allowPeers[p]
		if !ok {
			return false
		}
	}

	if len(a.allowSubnets) > 0 {
		ip, err := manet.ToIP(addr)
		if err != nil {
			return false
		}

		for _, ipnet := range a.allowSubnets {
			if ipnet.Contains(ip) {
				return true
			}
		}
		return false
	}

	return true
}

func (a *ACLFilter) InterceptPeerDial(p peer.ID) (allow bool) {
	return true
}

func (a *ACLFilter) InterceptAddrDial(peer.ID, ma.Multiaddr) (allow bool) {
	return true
}

func (a *ACLFilter) InterceptAccept(cm network.ConnMultiaddrs) (allow bool) {
	if len(a.allowSubnets) > 0 {
		addr := cm.RemoteMultiaddr()
		ip, err := manet.ToIP(addr)
		if err != nil {
			return false
		}

		for _, ipnet := range a.allowSubnets {
			if ipnet.Contains(ip) {
				return true
			}
		}
		return false
	}
	return true
}

func (a *ACLFilter) InterceptSecured(di network.Direction, p peer.ID, cm network.ConnMultiaddrs) (allow bool) {
	if di == network.DirOutbound {
		return true
	}
	if len(a.allowPeers) > 0 {
		_, ok := a.allowPeers[p]
		if !ok {
			return false
		}
	}
	return true
}

func (a *ACLFilter) InterceptUpgraded(network.Conn) (allow bool, reason control.DisconnectReason) {
	return true, 0
}
