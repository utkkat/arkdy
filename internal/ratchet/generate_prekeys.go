package ratchet

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"

	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/libsignal/util/optional"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func (e *Engine) Bundle(ctx context.Context, remote string) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if bundle, ok := e.bundles[remote]; ok {
		return bytesCopy(bundle), nil
	}
	opk := e.store.prekeys[e.nextPreKey]
	if opk == nil {
		return nil, fmt.Errorf("local prekey pool exhausted; restart the client")
	}
	spk := e.store.signed[1]
	sig := spk.Signature()
	data, err := structpb.NewStruct(map[string]any{
		"registration": e.store.registration, "opk_id": opk.ID().Value, "spk_id": spk.ID(),
		"ik":        base64.StdEncoding.EncodeToString(e.store.ik.PublicKey().Serialize()),
		"opk":       base64.StdEncoding.EncodeToString(opk.KeyPair().PublicKey().Serialize()),
		"spk":       base64.StdEncoding.EncodeToString(spk.KeyPair().PublicKey().Serialize()),
		"signature": base64.StdEncoding.EncodeToString(sig[:]),
	})
	if err != nil {
		return nil, err
	}
	encoded, err := proto.Marshal(data)
	if err != nil {
		return nil, err
	}
	e.bundles[remote] = encoded
	e.nextPreKey++
	return bytesCopy(encoded), nil
}

func bytesCopy(value []byte) []byte { return append([]byte(nil), value...) }

func (e *Engine) Establish(ctx context.Context, remote string, encoded []byte) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	builder, address := e.builder(remote)
	if _, ok := e.store.sessions[address.String()]; ok {
		return nil
	}
	bundle, err := decodeBundle(encoded)
	if err != nil {
		return err
	}
	return builder.ProcessBundle(ctx, bundle)
}

func decodeBundle(encoded []byte) (*prekey.Bundle, error) {
	var data structpb.Struct
	if err := proto.Unmarshal(encoded, &data); err != nil {
		return nil, fmt.Errorf("decode prekey bundle: %w", err)
	}
	number := func(name string) (uint32, error) {
		field := data.Fields[name]
		if field == nil {
			return 0, fmt.Errorf("missing bundle field %s", name)
		}
		n, ok := field.Kind.(*structpb.Value_NumberValue)
		if !ok || math.IsNaN(n.NumberValue) || n.NumberValue < 0 || n.NumberValue > math.MaxUint32 || math.Trunc(n.NumberValue) != n.NumberValue {
			return 0, fmt.Errorf("invalid bundle number %s", name)
		}
		return uint32(n.NumberValue), nil
	}
	binaryField := func(name string, length int) ([]byte, error) {
		raw, err := base64.StdEncoding.DecodeString(data.Fields[name].GetStringValue())
		if err != nil || len(raw) != length {
			return nil, fmt.Errorf("invalid bundle field %s", name)
		}
		return raw, nil
	}
	registration, err := number("registration")
	if err != nil {
		return nil, err
	}
	opkID, err := number("opk_id")
	if err != nil {
		return nil, err
	}
	spkID, err := number("spk_id")
	if err != nil {
		return nil, err
	}
	keys := make([]ecc.ECPublicKeyable, 0, 3)
	for _, name := range []string{"ik", "opk", "spk"} {
		raw, err := binaryField(name, 33)
		if err != nil {
			return nil, err
		}
		key, err := ecc.DecodePoint(raw, 0)
		if err != nil {
			return nil, fmt.Errorf("decode %s: %w", name, err)
		}
		keys = append(keys, key)
	}
	signature, err := binaryField("signature", 64)
	if err != nil {
		return nil, err
	}
	return prekey.NewBundle(registration, 1, optional.NewOptionalUint32(opkID), spkID, keys[1], keys[2], [64]byte(signature), identity.NewKey(keys[0])), nil
}
