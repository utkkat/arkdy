package ratchet

import (
	"context"
	"crypto/aes"
	"fmt"

	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/session"
)

func (e *Engine) Decrypt(ctx context.Context, remote string, envelope []byte) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(envelope) < 2 {
		return nil, fmt.Errorf("truncated encrypted envelope")
	}
	builder, address := e.builder(remote)
	cipher := session.NewCipher(builder, address)
	var padded []byte
	var err error
	switch envelope[0] {
	case protocol.PREKEY_TYPE:
		structure, decodeErr := e.store.serializer.PreKeySignalMessage.Deserialize(envelope[1:])
		if decodeErr != nil {
			return nil, decodeErr
		}
		if len(structure.BaseKey) != 33 || len(structure.IdentityKey) != 33 {
			return nil, fmt.Errorf("invalid prekey message keys")
		}
		if _, err := e.parseSignal(structure.Message); err != nil {
			return nil, err
		}
		message, decodeErr := protocol.NewPreKeySignalMessageFromStruct(structure, e.store.serializer.PreKeySignalMessage, e.store.serializer.SignalMessage)
		if decodeErr != nil {
			return nil, decodeErr
		}
		previousIdentity := e.store.identities[address.String()]
		padded, err = cipher.DecryptMessage(ctx, message)
		if err != nil {
			if previousIdentity == nil {
				delete(e.store.identities, address.String())
			} else {
				e.store.identities[address.String()] = previousIdentity
			}
		}
	case protocol.WHISPER_TYPE:
		message, decodeErr := e.parseSignal(envelope[1:])
		if decodeErr != nil {
			return nil, decodeErr
		}
		padded, err = cipher.Decrypt(ctx, message)
	default:
		return nil, fmt.Errorf("unknown encrypted message type %d", envelope[0])
	}
	if err != nil {
		return nil, fmt.Errorf("decrypt message: %w", err)
	}
	return Unpad(padded)
}

func (e *Engine) parseSignal(data []byte) (*protocol.SignalMessage, error) {
	structure, err := e.store.serializer.SignalMessage.Deserialize(data)
	if err != nil {
		return nil, err
	}
	if len(structure.RatchetKey) != 33 || len(structure.CipherText) == 0 || len(structure.CipherText)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid ratchet key or ciphertext length")
	}
	return protocol.NewSignalMessageFromStruct(structure, e.store.serializer.SignalMessage)
}
