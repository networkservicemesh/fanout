package fanout

import (
	"context"
	"net"
	"testing"

	"github.com/coredns/coredns/plugin/test"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// TestClientRequestTCRetriesTCP verifies the Telnyx fix for INFRA-798: when an
// upstream replies with a truncated (TC=1) UDP response, the client must
// retry the same query over TCP and return the full answer instead of
// relaying the truncated reply.
func TestClientRequestTCRetriesTCP(t *testing.T) {
	answers := []dns.RR{
		&dns.TXT{Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300}, Txt: []string{string(make([]byte, 200))}},
		&dns.TXT{Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300}, Txt: []string{string(make([]byte, 200))}},
		&dns.TXT{Hdr: dns.RR_Header{Name: "example.com.", Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 300}, Txt: []string{string(make([]byte, 200))}},
	}
	udpHandler := dns.HandlerFunc(func(w dns.ResponseWriter, q *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(q)
		m.Truncated = true // TC=1, no answers
		_ = w.WriteMsg(m)
	})
	tcpHandler := dns.HandlerFunc(func(w dns.ResponseWriter, q *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(q)
		m.Answer = answers
		_ = w.WriteMsg(m)
	})

	// Upstream serving UDP and TCP on the same address.
	lTCP, err := net.Listen(TCP, "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lTCP.Close()
	pcplUDP, err := net.ListenPacket("udp", lTCP.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer pcplUDP.Close()
	go dns.ActivateAndServe(nil, pcplUDP, udpHandler)
	go dns.ActivateAndServe(lTCP, nil, tcpHandler)

	addr := lTCP.Addr().String()

	q := new(dns.Msg)
	q.SetQuestion("example.com.", dns.TypeTXT)
	req := &request.Request{W: &test.ResponseWriter{}, Req: q} // UDP ResponseWriter

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // ensure the ctx-watcher goroutines in Request/requestTCP exit

	c := NewClient(addr, UDP)
	resp, err := c.Request(ctx, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.Truncated {
		t.Fatalf("expected TC to be resolved via TCP retry, got truncated reply")
	}
	if len(resp.Answer) != len(answers) {
		t.Fatalf("expected %d answers over TCP retry, got %d", len(answers), len(resp.Answer))
	}
}
