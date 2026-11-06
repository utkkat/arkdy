package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	ma "github.com/multiformats/go-multiaddr"
)

const defaultRelayConnectTimeout = 15 * time.Second

type HostConfig struct {
	ListenAddrs         []string
	RelayAddr           string
	RelayConnectTimeout time.Duration
}

func NewHost(ctx context.Context, cfg HostConfig) (host.Host, error) {
	if ctx == nil {
		return nil, fmt.Errorf("host context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("initialize host: %w", err)
	}
	if cfg.RelayConnectTimeout < 0 {
		return nil, fmt.Errorf("relay connect timeout must not be negative")
	}
	if cfg.RelayConnectTimeout == 0 {
		cfg.RelayConnectTimeout = defaultRelayConnectTimeout
	}

	listenAddrs := cfg.ListenAddrs
	if len(listenAddrs) == 0 {
		listenAddrs = []string{
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic-v1",
		}
	}
	addrs := make([]ma.Multiaddr, 0, len(listenAddrs))
	for _, addr := range listenAddrs {
		parsed, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("parse listen address %q: %w", addr, err)
		}
		addrs = append(addrs, parsed)
	}

	var relayInfo *peer.AddrInfo
	if cfg.RelayAddr != "" {
		var err error
		relayInfo, err = peer.AddrInfoFromString(cfg.RelayAddr)
		if err != nil {
			return nil, fmt.Errorf("parse relay address: %w", err)
		}
		if len(relayInfo.Addrs) == 0 {
			return nil, fmt.Errorf("relay address must include a transport address")
		}
	}

	ek, err := GenerateEphemeralKey()
	if err != nil {
		return nil, fmt.Errorf("generate host identity: %w", err)
	}

	manager, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		return nil, fmt.Errorf("create connection manager: %w", err)
	}
	h, err := libp2p.New(
		libp2p.Identity(ek.priv),
		libp2p.ListenAddrs(addrs...),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(manager),
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		_ = manager.Close()
		return nil, fmt.Errorf("create libp2p host: %w", err)
	}
	ready := false
	defer func() {
		if !ready {
			_ = h.Close()
		}
	}()

	if relayInfo != nil {
		connectCtx, cancel := context.WithTimeout(ctx, cfg.RelayConnectTimeout)
		defer cancel()
		if err := h.Connect(connectCtx, *relayInfo); err != nil {
			return nil, fmt.Errorf("connect to relay %s: %w", relayInfo.ID, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("initialize host: %w", err)
	}

	ready = true
	return h, nil
}
