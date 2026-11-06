package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio"
)

func SetMessageHandler(ctx context.Context, h host.Host, receive func(context.Context, peer.ID, []byte) error) {
	setHandler(ctx, h, protocolID, func(ctx context.Context, remote peer.ID, payload []byte) ([]byte, error) {
		if len(payload) == 0 {
			return nil, fmt.Errorf("empty envelope")
		}
		return nil, receive(ctx, remote, payload)
	})
}

func SetPreKeyBundleHandler(ctx context.Context, h host.Host, bundle func(context.Context, peer.ID) ([]byte, error)) {
	setHandler(ctx, h, bundleProtocolID, func(ctx context.Context, remote peer.ID, payload []byte) ([]byte, error) {
		if len(payload) != 0 {
			return nil, fmt.Errorf("bundle request must be empty")
		}
		return bundle(ctx, remote)
	})
}

func setHandler(ctx context.Context, h host.Host, id protocol.ID, handle func(context.Context, peer.ID, []byte) ([]byte, error)) {
	h.SetStreamHandler(id, func(stream network.Stream) {
		defer stream.Close()
		requestCtx, cancel := context.WithTimeout(ctx, ioTimeout)
		defer cancel()
		stop := context.AfterFunc(requestCtx, func() { _ = stream.Reset() })
		defer stop()
		if err := stream.SetDeadline(time.Now().Add(ioTimeout)); err != nil {
			_ = stream.Reset()
			return
		}
		payload, err := readFrame(msgio.NewVarintReaderSize(stream, maxEnvelope))
		if err != nil {
			_ = stream.Reset()
			return
		}
		response, err := handle(requestCtx, stream.Conn().RemotePeer(), payload)
		if err != nil {
			_ = stream.Reset()
			return
		}
		frame, err := marshalFrame(response)
		if err != nil {
			_ = stream.Reset()
			return
		}
		if err := msgio.NewVarintWriter(stream).WriteMsg(frame); err != nil {
			_ = stream.Reset()
		}
	})
}
