package protocol

import (
	"io"
	"net"

	"github.com/txthinking/socks5"
)

func Socks5Handler(bs *BufReaderStream) {
	if err := socks5Negotiate(bs); err != nil {
		Log.Error(err)
		return
	}

	conn, err := socks5RequestConnect(bs)
	if err != nil {
		Log.Error(err)
		return
	}
	defer conn.Close()

	if err := tunneling(bs, conn); err != nil {
		Log.Error(err)
	}
}

func IsSocks5(v byte) bool {
	return v == socks5.Ver
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

func socks5RequestConnect(rw io.ReadWriter) (net.Conn, error) {
	r, err := socks5.NewRequestFrom(rw)
	if err != nil {
		return nil, err
	}

	if r.Cmd != socks5.CmdConnect {
		if e := replyErr(r, rw, socks5.RepCommandNotSupported); err != nil {
			return nil, e
		}
		return nil, socks5.ErrUnsupportCmd
	}

	conn, err := net.Dial("tcp", r.Address())
	if err != nil {
		if e := replyErr(r, rw, socks5.RepHostUnreachable); err != nil {
			return nil, e
		}
		return nil, err
	}

	a, addr, port, err := socks5.ParseAddress(conn.LocalAddr().String())
	if err != nil {
		if e := replyErr(r, rw, socks5.RepHostUnreachable); err != nil {
			return nil, e
		}
		return nil, err
	}

	p := socks5.NewReply(socks5.RepSuccess, a, addr, port)
	if _, err := p.WriteTo(rw); err != nil {
		return nil, err
	}

	Log.Infof("socks5 proxying for %s", r.Address())
	return conn, nil
}

func replyErr(req *socks5.Request, rw io.ReadWriter, rep byte) error {
	var p *socks5.Reply
	if req.Atyp == socks5.ATYPIPv4 || req.Atyp == socks5.ATYPDomain {
		p = socks5.NewReply(rep, socks5.ATYPIPv4, []byte{0x00, 0x00, 0x00, 0x00}, []byte{0x00, 0x00})
	} else {
		p = socks5.NewReply(rep, socks5.ATYPIPv6, []byte(net.IPv6zero), []byte{0x00, 0x00})
	}
	_, err := p.WriteTo(rw)
	return err
}
