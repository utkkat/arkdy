package transport

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-msgio"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	protocolID       = protocol.ID("/arkdy/direct/1.0.0")
	bundleProtocolID = protocol.ID("/arkdy/prekeys/1.0.0")
	maxEnvelope      = 20 * 1024
	ioTimeout        = 15 * time.Second
	dialTimeout      = 30 * time.Second
)

func DirectSend(ctx context.Context, h host.Host, target peer.ID, envelope []byte) error {
	if len(envelope) == 0 {
		return fmt.Errorf("empty envelope")
	}
	ack, err := exchange(ctx, h, target, protocolID, envelope)
	if err != nil {
		return err
	}
	if len(ack) != 0 {
		return fmt.Errorf("unexpected delivery acknowledgement")
	}
	return nil
}

func RequestPreKeyBundle(ctx context.Context, h host.Host, target peer.ID) ([]byte, error) {
	return exchange(ctx, h, target, bundleProtocolID, nil)
}

func exchange(ctx context.Context, h host.Host, target peer.ID, id protocol.ID, payload []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	frame, err := marshalFrame(payload)
	if err != nil {
		return nil, err
	}
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()
	stream, err := h.NewStream(dialCtx, target, id)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", target, err)
	}
	defer stream.Close()
	stop := context.AfterFunc(ctx, func() { _ = stream.Reset() })
	defer stop()
	deadline := time.Now().Add(ioTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := stream.SetDeadline(deadline); err != nil {
		_ = stream.Reset()
		return nil, err
	}
	if err := msgio.NewVarintWriter(stream).WriteMsg(frame); err != nil {
		_ = stream.Reset()
		return nil, fmt.Errorf("write envelope: %w", err)
	}
	response, err := readFrame(msgio.NewVarintReaderSize(stream, maxEnvelope))
	if err != nil {
		_ = stream.Reset()
		return nil, fmt.Errorf("read response: %w", err)
	}
	return response, nil
}

func marshalFrame(payload []byte) ([]byte, error) {
	msg := wrapperspb.Bytes(payload)
	if proto.Size(msg) > maxEnvelope {
		return nil, fmt.Errorf("envelope exceeds %d bytes", maxEnvelope)
	}
	return proto.Marshal(msg)
}

func readFrame(reader msgio.ReadCloser) ([]byte, error) {
	frame, err := reader.ReadMsg()
	if err != nil {
		return nil, err
	}
	defer reader.ReleaseMsg(frame)
	var msg wrapperspb.BytesValue
	if err := proto.Unmarshal(frame, &msg); err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}
	return msg.Value, nil
}
