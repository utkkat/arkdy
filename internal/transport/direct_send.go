// NOT FINISHED YET!

package transport

import (
	// context.Background задаёт контекст открытия потока; WithTimeout пригодится для dialTimeout.
	// https://pkg.go.dev/context
	"context"
	// time.Second задаёт длительности подключения и операций ввода-вывода.
	// https://pkg.go.dev/time#Duration
	"time"

	// host.Host предоставляет NewStream и доступ к Peerstore.
	// https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/host#Host
	"github.com/libp2p/go-libp2p/core/host"
	// peer.ID обозначает узел; AddrInfoFromP2pAddr извлекает ID и адреса из multiaddr.
	// https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/peer
	"github.com/libp2p/go-libp2p/core/peer"
	// Peerstore хранит известные адреса; PermanentAddrTTL задаёт срок их хранения.
	// https://pkg.go.dev/github.com/libp2p/go-libp2p@v0.50.0/core/peerstore
	"github.com/libp2p/go-libp2p/core/peerstore"
	// ma.NewMultiaddr разбирает адрес вида /ip4/.../tcp/.../p2p/..., а не строку PeerID.
	// https://pkg.go.dev/github.com/multiformats/go-multiaddr@v0.16.1#NewMultiaddr
	ma "github.com/multiformats/go-multiaddr"
)

// Выбранные импорты для будущего кодирования и framing; текущая функция ещё пишет сырые байты:
// import "github.com/libp2p/go-msgio" // NewVarintWriter: https://pkg.go.dev/github.com/libp2p/go-msgio#NewVarintWriter
// import "google.golang.org/protobuf/proto" // Marshal: https://pkg.go.dev/google.golang.org/protobuf/proto#Marshal
// Подключать одновременно с receive_direct.go после определения .proto-схемы конверта.

const (
	protocolID  = "/arkdy/direct/1.0.0"
	maxEnvelope = 20 * 1024
	ioTimeout   = 15 * time.Second
	dialTimeout = 30 * time.Second
)

func directSend(h host.Host, targetPeerID peer.ID, message string) error {
	// TODO: адрес получать отдельно от PeerID; string(targetPeerID) не является multiaddr.
	targetAddr, err := ma.NewMultiaddr(string(targetPeerID))
	if err != nil {
		return err
	}
	info, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return err
	}
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	stream, err := h.NewStream(context.Background(), info.ID, protocolID)
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
