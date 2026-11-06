package ratchet

import (
	"context"
	"fmt"

	"go.mau.fi/libsignal/session"
)

func (e *Engine) Encrypt(ctx context.Context, remote string, plaintext []byte) ([]byte, error) {
	padded, err := Pad(plaintext)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	builder, address := e.builder(remote)
	if _, ok := e.store.sessions[address.String()]; !ok {
		return nil, fmt.Errorf("no session for %s", remote)
	}
	message, err := session.NewCipher(builder, address).Encrypt(ctx, padded)
	if err != nil {
		return nil, fmt.Errorf("encrypt message: %w", err)
	}
	return append([]byte{byte(message.Type())}, message.Serialize()...), nil
}
