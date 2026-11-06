package core

import (
	"context"
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/utkkat/arkdy/internal/ratchet"
	"github.com/utkkat/arkdy/internal/transport"
)

const MaxMessageSize = ratchet.MaxMessageSize

type Message struct {
	From peer.ID
	Text string
}

type Client struct {
	host      host.Host
	ratchet   *ratchet.Engine
	ctx       context.Context
	cancel    context.CancelFunc
	messages  chan Message
	sendMu    sync.Mutex
	closeOnce sync.Once
	closeErr  error
}

func NewClient(ctx context.Context, cfg transport.HostConfig) (*Client, error) {
	h, err := transport.NewHost(ctx, cfg)
	if err != nil {
		return nil, err
	}
	engine, err := ratchet.New()
	if err != nil {
		_ = h.Close()
		return nil, err
	}
	clientCtx, cancel := context.WithCancel(ctx)
	c := &Client{host: h, ratchet: engine, ctx: clientCtx, cancel: cancel, messages: make(chan Message, 64)}
	transport.SetPreKeyBundleHandler(clientCtx, h, func(ctx context.Context, remote peer.ID) ([]byte, error) {
		return engine.Bundle(ctx, remote.String())
	})
	transport.SetMessageHandler(clientCtx, h, func(ctx context.Context, remote peer.ID, envelope []byte) error {
		plaintext, err := engine.Decrypt(ctx, remote.String(), envelope)
		if err != nil {
			return err
		}
		if len(plaintext) == 0 || !utf8.Valid(plaintext) {
			return fmt.Errorf("invalid text message")
		}
		select {
		case c.messages <- Message{From: remote, Text: string(plaintext)}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	return c, nil
}

func (c *Client) ID() peer.ID { return c.host.ID() }

func (c *Client) Addresses() []string {
	info := peer.AddrInfo{ID: c.host.ID(), Addrs: c.host.Addrs()}
	addrs, _ := peer.AddrInfoToP2pAddrs(&info)
	result := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		result = append(result, addr.String())
	}
	return result
}

func (c *Client) Messages() <-chan Message { return c.messages }

func (c *Client) Connect(ctx context.Context, address string) (peer.ID, error) {
	if err := c.ctx.Err(); err != nil {
		return "", err
	}
	info, err := peer.AddrInfoFromString(address)
	if err != nil {
		return "", fmt.Errorf("invalid peer address: %w", err)
	}
	if len(info.Addrs) == 0 {
		return "", fmt.Errorf("peer address must include a transport address")
	}
	if info.ID == c.ID() {
		return "", fmt.Errorf("cannot connect to yourself")
	}
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	stop := context.AfterFunc(c.ctx, cancel)
	defer stop()
	if err := c.host.Connect(connectCtx, *info); err != nil {
		return "", fmt.Errorf("connect: %w", err)
	}
	return info.ID, nil
}

func (c *Client) Send(ctx context.Context, target peer.ID, text string) error {
	if len(text) == 0 || !utf8.ValidString(text) {
		return fmt.Errorf("message must be nonempty UTF-8 text")
	}
	if len(text) > MaxMessageSize {
		return fmt.Errorf("message exceeds %d bytes", MaxMessageSize)
	}
	if err := target.Validate(); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	if target == c.ID() {
		return fmt.Errorf("cannot send to yourself")
	}
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if err := c.ctx.Err(); err != nil {
		return err
	}
	sendCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(c.ctx, cancel)
	defer stop()
	if !c.ratchet.HasSession(target.String()) {
		bundle, err := transport.RequestPreKeyBundle(sendCtx, c.host, target)
		if err != nil {
			return fmt.Errorf("fetch prekeys: %w", err)
		}
		if err := c.ratchet.Establish(sendCtx, target.String(), bundle); err != nil {
			return fmt.Errorf("establish session: %w", err)
		}
	}
	envelope, err := c.ratchet.Encrypt(sendCtx, target.String(), []byte(text))
	if err != nil {
		return err
	}
	return transport.DirectSend(sendCtx, c.host, target, envelope)
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.cancel()
		c.closeErr = c.host.Close()
	})
	return c.closeErr
}
