package transport

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

func main() {
	hostGen()
}

func hostGen() (host.Host, error) {
	ek, err := GenerateEphemeralKey()
	if err != nil {
		panic(err)
	}

	connmgr, err := connmgr.NewConnManager(100, 400, connmgr.WithGracePeriod(time.Minute))
	if err != nil {
		panic(err)
	}
	h, err := libp2p.New(
		libp2p.Identity(ek.priv),
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",
			"/ip4/0.0.0.0/udp/9000/quic-v1",
		),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		libp2p.DefaultTransports,
		libp2p.ConnectionManager(connmgr),
		libp2p.NATPortMap(),
		libp2p.EnableHolePunching(),
		libp2p.EnableNATService(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		panic(err)
	}
	log.Printf("my hosts ID is %s\n", h.ID())

	relayAddr := "/ip4/RELAY_IP/tcp/4001/p2p/RELAY_PEER_ID" //TODO: Replace with act.

	relayInfo, err := peer.AddrInfoFromString(relayAddr)
	if err != nil {
		log.Printf("Invalid relay address: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	if err := h.Connect(ctx, *relayInfo); err != nil {
		log.Printf("Failed to connect to relay: %v", err)
		return nil, err
	}

	log.Printf("Connected to relay %s", relayInfo.ID)
	return h, nil
}
