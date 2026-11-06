package core_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/utkkat/arkdy/internal/core"
	"github.com/utkkat/arkdy/internal/transport"
)

func newClient(t *testing.T, ctx context.Context) *core.Client {
	t.Helper()
	c, err := core.NewClient(ctx, transport.HostConfig{ListenAddrs: []string{"/ip4/127.0.0.1/tcp/0"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Error(err)
		}
	})
	return c
}

func receive(t *testing.T, ctx context.Context, recipient, sender *core.Client, want string) {
	t.Helper()
	select {
	case msg := <-recipient.Messages():
		if msg.From != sender.ID() || msg.Text != want {
			t.Fatalf("unexpected message from %s: %q", msg.From, msg.Text)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for message")
	}
}

func TestLocalConversation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, bob := newClient(t, ctx), newClient(t, ctx)
	id, err := alice.Connect(ctx, bob.Addresses()[0])
	if err != nil || id != bob.ID() {
		t.Fatalf("Connect = %s, %v", id, err)
	}
	for _, text := range []string{"Привет 👋", " spaces preserved ", strings.Repeat("x", core.MaxMessageSize)} {
		if err := alice.Send(ctx, id, text); err != nil {
			t.Fatal(err)
		}
		receive(t, ctx, bob, alice, text)
	}
	if err := bob.Send(ctx, alice.ID(), "Ответ"); err != nil {
		t.Fatal(err)
	}
	receive(t, ctx, alice, bob, "Ответ")
	if err := alice.Send(ctx, id, strings.Repeat("x", core.MaxMessageSize+1)); err == nil {
		t.Fatal("oversized message accepted")
	}
	if err := alice.Send(ctx, id, ""); err == nil {
		t.Fatal("empty message accepted")
	}
	if err := alice.Send(ctx, id, "after error"); err != nil {
		t.Fatal(err)
	}
	receive(t, ctx, bob, alice, "after error")
}

func TestConcurrentFirstMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, bob := newClient(t, ctx), newClient(t, ctx)
	if _, err := alice.Connect(ctx, bob.Addresses()[0]); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, pair := range [][2]*core.Client{{alice, bob}, {bob, alice}} {
		wg.Add(1)
		go func(from, to *core.Client) {
			defer wg.Done()
			if err := from.Send(ctx, to.ID(), "simultaneous"); err != nil {
				t.Error(err)
			}
		}(pair[0], pair[1])
	}
	wg.Wait()
	receive(t, ctx, bob, alice, "simultaneous")
	receive(t, ctx, alice, bob, "simultaneous")
	if err := alice.Send(ctx, bob.ID(), "follow-up"); err != nil {
		t.Fatal(err)
	}
	receive(t, ctx, bob, alice, "follow-up")
}

func TestClientErrorsAndClose(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	alice, bob := newClient(t, ctx), newClient(t, ctx)
	for _, address := range []string{"bad", "/p2p/" + bob.ID().String(), alice.Addresses()[0]} {
		if _, err := alice.Connect(ctx, address); err == nil {
			t.Fatalf("accepted address %q", address)
		}
	}
	canceled, stop := context.WithCancel(ctx)
	stop()
	if err := alice.Send(canceled, bob.ID(), "canceled"); err == nil {
		t.Fatal("send ignored cancellation")
	}
	if err := alice.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := alice.Connect(ctx, bob.Addresses()[0]); err == nil {
		t.Fatal("closed client connected")
	}
	if err := alice.Send(ctx, bob.ID(), "closed"); err == nil {
		t.Fatal("closed client sent")
	}
}
