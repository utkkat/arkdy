package transport_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-msgio"
	"github.com/utkkat/arkdy/internal/transport"
)

func connectedHosts(t *testing.T, ctx context.Context) (host.Host, host.Host) {
	t.Helper()
	var hosts []host.Host
	for i := 0; i < 2; i++ {
		h, err := transport.NewHost(ctx, transport.HostConfig{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := h.Close(); err != nil {
				t.Error(err)
			}
		})
		hosts = append(hosts, h)
	}
	if err := hosts[0].Connect(ctx, peer.AddrInfo{ID: hosts[1].ID(), Addrs: hosts[1].Addrs()}); err != nil {
		t.Fatal(err)
	}
	return hosts[0], hosts[1]
}

func TestDirectSendAcknowledgement(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, bob := connectedHosts(t, ctx)
	received := make(chan []byte, 1)
	transport.SetMessageHandler(ctx, bob, func(_ context.Context, remote peer.ID, payload []byte) error {
		if remote != alice.ID() {
			return fmt.Errorf("wrong sender")
		}
		received <- payload
		return nil
	})
	want := []byte{0, 1, 2, 0xff, '\n'}
	if err := transport.DirectSend(ctx, alice, bob.ID(), want); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-received:
		if !bytes.Equal(got, want) {
			t.Fatalf("payload = %x", got)
		}
	default:
		t.Fatal("sender got acknowledgement before handler accepted the payload")
	}
	transport.SetMessageHandler(ctx, bob, func(context.Context, peer.ID, []byte) error { return fmt.Errorf("rejected") })
	if err := transport.DirectSend(ctx, alice, bob.ID(), want); err == nil {
		t.Fatal("rejection reported as success")
	}
}

func TestMalformedFrames(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, bob := connectedHosts(t, ctx)
	received := make(chan struct{}, 1)
	transport.SetMessageHandler(ctx, bob, func(context.Context, peer.ID, []byte) error { received <- struct{}{}; return nil })
	for _, oversized := range []bool{false, true} {
		stream, err := alice.NewStream(ctx, bob.ID(), "/arkdy/direct/1.0.0")
		if err != nil {
			t.Fatal(err)
		}
		_ = stream.SetDeadline(time.Now().Add(2 * time.Second))
		if oversized {
			_, err = stream.Write(binary.AppendUvarint(nil, 20*1024+1))
		} else {
			err = msgio.NewVarintWriter(stream).WriteMsg([]byte{0xff})
		}
		if err != nil {
			t.Fatal(err)
		}
		var response [1]byte
		if _, err := stream.Read(response[:]); err == nil {
			t.Fatal("malformed frame acknowledged")
		}
		_ = stream.Close()
	}
	select {
	case <-received:
		t.Fatal("malformed frame reached receiver")
	default:
	}
	if err := transport.DirectSend(ctx, alice, bob.ID(), []byte("still alive")); err != nil {
		t.Fatal(err)
	}
	if err := transport.DirectSend(ctx, alice, bob.ID(), make([]byte, 20*1024)); err == nil {
		t.Fatal("oversized outgoing envelope accepted")
	}
}

func TestDirectSendCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, bob := connectedHosts(t, ctx)
	started, release := make(chan struct{}), make(chan struct{})
	defer close(release)
	transport.SetMessageHandler(ctx, bob, func(context.Context, peer.ID, []byte) error {
		close(started)
		<-release
		return nil
	})
	sendCtx, stop := context.WithCancel(ctx)
	defer stop()
	done := make(chan error, 1)
	go func() { done <- transport.DirectSend(sendCtx, alice, bob.ID(), []byte("waiting")) }()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("handler did not start")
	}
	stop()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled send succeeded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("send ignored cancellation")
	}
}
