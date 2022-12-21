package protocol

import (
	"io"
	"net"

	"github.com/txthinking/socks5"
)

func IsSocks5(v byte) bool {
	return v == socks5.Ver
}

func (p *ProxyService) socks5Handler(bs *BufReaderStream) {
	if err := socks5Negotiate(bs); shouldLogError(err) {
		Log.Error(err)
		return
	}

	if err := p.socks5RequestConnect(bs); shouldLogError(err) {
		Log.Warn(err)
	}
}

func (p *ProxyService) socks5RequestConnect(bs *BufReaderStream) error {
	r, err := socks5.NewRequestFrom(bs.Reader)
	if err != nil {
		return err
	}

	if r.Cmd != socks5.CmdConnect {
		if e := replyErr(r, bs, socks5.RepCommandNotSupported); err != nil {
			return e
		}
		return socks5.ErrUnsupportCmd
	}

	if p.isP2PHttp(r.Address()) {
		a, addr, port, err := socks5.ParseAddress(r.Address())
		if err != nil {
			if e := replyErr(r, bs, socks5.RepHostUnreachable); err != nil {
				return e
			}
			return err
		}

		reply := socks5.NewReply(socks5.RepSuccess, a, addr, port)
		if _, err := reply.WriteTo(bs); err != nil {
			return err
		}
		p.p2phttpHandler(bs, nil)
		return nil
	}

	conn, err := net.Dial("tcp", r.Address())
	if err != nil {
		if e := replyErr(r, bs, socks5.RepHostUnreachable); err != nil {
			return e
		}
		return err
	}

	defer conn.Close()
	a, addr, port, err := socks5.ParseAddress(conn.LocalAddr().String())
	if err != nil {
		if e := replyErr(r, bs, socks5.RepHostUnreachable); err != nil {
			return e
		}
		return err
	}

	reply := socks5.NewReply(socks5.RepSuccess, a, addr, port)
	if _, err := reply.WriteTo(bs); err != nil {
		return err
	}

	return tunneling(conn, bs)
}

func socks5Negotiate(bs *BufReaderStream) error {
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

func replyErr(req *socks5.Request, rw io.ReadWriter, rep byte) error {
	var reply *socks5.Reply
	if req.Atyp == socks5.ATYPIPv4 || req.Atyp == socks5.ATYPDomain {
		reply = socks5.NewReply(rep, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
	} else {
		reply = socks5.NewReply(rep, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
	}
	_, err := reply.WriteTo(rw)
	return err
}
