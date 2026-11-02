package transport	


import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/host"

	ma "github.com/multiformats/go-multiaddr"
)

func relaySend(h host.Host, targetPeerID peer.ID, message string) error {
	stream, err := h.NewStream(context.Background(), targetPeerID, protocol) // protocol from serv
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