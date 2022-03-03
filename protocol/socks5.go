package protocol

import (
	"io"
	"net"

	"github.com/txthinking/socks5"
)

func Socks5Handler(bs *BufStream) {
	if err := socks5Negotiate(bs); err != nil {
		log.Error(err)
		return
	}

	conn, err := socks5RequestConnect(bs.Stream)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	if err := tunneling(bs.Stream, conn); err != nil {
		log.Error(err)
	}
}

func IsSocks5(v byte) bool {
	return v == socks5.Ver
}

func socks5Negotiate(bs *BufStream) error {
	rq, err := socks5.NewNegotiationRequestFrom(bs.Reader)
	if err != nil {
		return err
	}

	for _, m := range rq.Methods {
		if m == socks5.MethodNone {
			rp := socks5.NewNegotiationReply(socks5.MethodNone)
			_, err = rp.WriteTo(bs)
			return err
		}
	}

	rp := socks5.NewNegotiationReply(socks5.MethodUnsupportAll)
	_, err = rp.WriteTo(bs)
	return err
}

func socks5RequestConnect(rw io.ReadWriter) (net.Conn, error) {
	r, err := socks5.NewRequestFrom(rw)
	if err != nil {
		return nil, err
	}

	if r.Cmd != socks5.CmdConnect {
		var p *socks5.Reply
		if r.Atyp == socks5.ATYPIPv4 || r.Atyp == socks5.ATYPDomain {
			p = socks5.NewReply(socks5.RepCommandNotSupported, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = socks5.NewReply(socks5.RepCommandNotSupported, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(rw); err != nil {
			return nil, err
		}
		return nil, socks5.ErrUnsupportCmd
	}

	rc, err := net.Dial("tcp", r.Address())
	if err != nil {
		var p *socks5.Reply
		if r.Atyp == socks5.ATYPIPv4 || r.Atyp == socks5.ATYPDomain {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(rw); err != nil {
			return nil, err
		}
		return nil, err
	}

	a, addr, port, err := socks5.ParseAddress(rc.LocalAddr().String())
	if err != nil {
		var p *socks5.Reply
		if r.Atyp == socks5.ATYPIPv4 || r.Atyp == socks5.ATYPDomain {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
		} else {
			p = socks5.NewReply(socks5.RepHostUnreachable, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
		}
		if _, err := p.WriteTo(rw); err != nil {
			return nil, err
		}
		return nil, err
	}
	p := socks5.NewReply(socks5.RepSuccess, a, addr, port)
	if _, err := p.WriteTo(rw); err != nil {
		return nil, err
	}

	return rc, nil
}
