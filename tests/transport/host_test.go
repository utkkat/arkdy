package transport_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/utkkat/arkdy/internal/transport"
)

func TestNewHost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h, err := transport.NewHost(ctx, transport.HostConfig{
		ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"},
	})
	if err != nil {
		t.Fatalf("NewHost() failed: %v", err)
	}
	if h == nil {
		t.Fatal("NewHost() returned a nil host")
	}
	t.Cleanup(func() {
		if err := h.Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})

	if err := h.ID().Validate(); err != nil {
		t.Fatalf("host has an invalid peer ID: %v", err)
	}
	key := h.Peerstore().PrivKey(h.ID())
	if key == nil {
		t.Fatal("host identity has no private key")
	}
	wantID, err := peer.IDFromPrivateKey(key)
	if err != nil {
		t.Fatalf("derive peer ID from host key: %v", err)
	}
	if h.ID() != wantID {
		t.Errorf("host ID = %s, want %s", h.ID(), wantID)
	}

	addrs := h.Addrs()
	if len(addrs) != 1 {
		t.Fatalf("host addresses = %v, want one loopback TCP address", addrs)
	}
	ip, err := addrs[0].ValueForProtocol(ma.P_IP4)
	if err != nil || ip != "127.0.0.1" {
		t.Errorf("listen IP = %q, error = %v; want 127.0.0.1", ip, err)
	}
	port, err := addrs[0].ValueForProtocol(ma.P_TCP)
	if err != nil {
		t.Fatalf("listen address has no TCP port: %v", err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		t.Errorf("listen TCP port = %q, want an assigned port in 1..65535", port)
	}
}
