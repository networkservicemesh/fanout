package fanout

import (
	"context"
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func TestUseRequestSizeOnConn(t *testing.T) {
	s := newServer("udp", func(w dns.ResponseWriter, r *dns.Msg) {
		msg := dns.Msg{
			Answer: []dns.RR{
				makeRecordA("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk. 3600	IN	A 10.0.0.1"),
				makeRecordA("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk. 3600	IN	A 10.0.0.1"),
				makeRecordA("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk.abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijk. 3600	IN	A 10.0.0.1"),
			},
		}
		msg.SetReply(r)
		logErrIfNotNil(w.WriteMsg(&msg))
	})
	defer s.close()
	c := NewClient(s.addr, "udp")
	req := new(dns.Msg)
	req.SetEdns0(dns.DefaultMsgSize, false)
	req.SetQuestion(testQuery, dns.TypeA)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	d, err := c.Request(ctx, &request.Request{W: &test.ResponseWriter{}, Req: req})
	require.Nil(t, err)
	require.Len(t, d.Answer, 3)
}
