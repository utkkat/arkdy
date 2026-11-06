package transport

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)


func relaySend(h host.Host, targetPeerID peer.ID, message string) error {
	stream, err := h.NewStream(context.Background(), targetPeerID, protocolID)
	if err != nil {
		return err
	}
	defer stream.Close()

	_, err = stream.Write([]byte(message))
	if err != nil {
		return err
	}

	return nil
}
