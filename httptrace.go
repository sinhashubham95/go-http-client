package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http/httptrace"
	"net/textproto"
)

func (c *Client) getHTTPTracer(ctx context.Context) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("get conn: %s", hostPort))
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("got conn: reused %t, was idle %t, idle duration millis %d",
					info.Reused, info.WasIdle, info.IdleTime.Milliseconds()))
			}
		},
		PutIdleConn: func(err error) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("put idle connection: %v", err))
			}
		},
		GotFirstResponseByte: func() {
			if c.l != nil {
				c.l(ctx, "got first response byte")
			}
		},
		Got100Continue: func() {
			if c.l != nil {
				c.l(ctx, "got 100 continue")
			}
		},
		Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("got 1xx response: code %d, header %v", code, header))
			}
			return nil
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("dns start: host %s", info.Host))
			}
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("dns done: err %v, is coalesced %t", info.Err, info.Coalesced))
			}
		},
		ConnectStart: func(network, addr string) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("connect start: network %s, address %s", network, addr))
			}
		},
		ConnectDone: func(network, addr string, err error) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("connect done: network %s, address %s, err %v", network, addr, err))
			}
		},
		TLSHandshakeStart: func() {
			if c.l != nil {
				c.l(ctx, "tls handshake start")
			}
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("tls handshake done: version %d, handshake complete: %t server name %s, err %v",
					state.Version, state.HandshakeComplete, state.ServerName, err))
			}
		},
		WroteHeaderField: func(_ string, _ []string) {},
		WroteHeaders: func() {
			if c.l != nil {
				c.l(ctx, "wrote headers")
			}
		},
		Wait100Continue: func() {
			if c.l != nil {
				c.l(ctx, "wait 100 continue")
			}
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if c.l != nil {
				c.l(ctx, fmt.Sprintf("wrote request: err %v", info.Err))
			}
		},
	}
}
