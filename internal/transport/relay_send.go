package transport

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

func RelaySend(ctx context.Context, h host.Host, target peer.ID, envelope []byte) error {
	return DirectSend(ctx, h, target, envelope)
}
