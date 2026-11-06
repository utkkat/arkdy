package ratchet_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/utkkat/arkdy/internal/ratchet"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/util/keyhelper"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func newEngine(t *testing.T) *ratchet.Engine {
	t.Helper()
	e, err := ratchet.New()
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestEncryptedConversation(t *testing.T) {
	ctx := context.Background()
	alice, bob := newEngine(t), newEngine(t)
	bundle, err := bob.Bundle(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if err := alice.Establish(ctx, "bob", bundle); err != nil {
		t.Fatal(err)
	}

	first, err := alice.Encrypt(ctx, "bob", []byte("Привет, Боб!"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(first, []byte("Привет, Боб!")) {
		t.Fatal("plaintext present in envelope")
	}
	serializer := serialize.NewProtoBufSerializer()
	forged, err := serializer.PreKeySignalMessage.Deserialize(first[1:])
	if err != nil {
		t.Fatal(err)
	}
	forgedIdentity, err := keyhelper.GenerateIdentityKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	forged.IdentityKey = forgedIdentity.PublicKey().Serialize()
	forgedEnvelope := append([]byte{protocol.PREKEY_TYPE}, serializer.PreKeySignalMessage.Serialize(forged)...)
	if _, err := bob.Decrypt(ctx, "alice", forgedEnvelope); err == nil {
		t.Fatal("forged first-use identity accepted")
	}
	if bob.HasSession("alice") {
		t.Fatal("forged first message created a session")
	}
	plain, err := bob.Decrypt(ctx, "alice", first)
	if err != nil || string(plain) != "Привет, Боб!" {
		t.Fatalf("first message = %q, %v", plain, err)
	}
	if _, err := bob.Decrypt(ctx, "alice", first); err == nil {
		t.Fatal("replayed message accepted")
	}
	charlie := newEngine(t)
	if err := charlie.Establish(ctx, "bob", bundle); err != nil {
		t.Fatal(err)
	}
	reused, err := charlie.Encrypt(ctx, "bob", []byte("reuse OPK"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bob.Decrypt(ctx, "charlie", reused); err == nil {
		t.Fatal("one-time prekey was reused by another session")
	}

	reply, err := bob.Encrypt(ctx, "alice", []byte("Привет, Алиса!"))
	if err != nil {
		t.Fatal(err)
	}
	plain, err = alice.Decrypt(ctx, "bob", reply)
	if err != nil || string(plain) != "Привет, Алиса!" {
		t.Fatalf("reply = %q, %v", plain, err)
	}

	var envelopes [][]byte
	for _, text := range []string{"one", "two", "three"} {
		envelope, err := alice.Encrypt(ctx, "bob", []byte(text))
		if err != nil {
			t.Fatal(err)
		}
		envelopes = append(envelopes, envelope)
	}
	tampered := bytes.Clone(envelopes[2])
	tampered[len(tampered)-1] ^= 1
	if _, err := bob.Decrypt(ctx, "alice", tampered); err == nil {
		t.Fatal("tampered message accepted")
	}
	for _, index := range []int{2, 0, 1} {
		plain, err := bob.Decrypt(ctx, "alice", envelopes[index])
		want := []string{"one", "two", "three"}[index]
		if err != nil || string(plain) != want {
			t.Fatalf("out-of-order message = %q, %v; want %q", plain, err, want)
		}
	}
}

func TestBundleValidationAndReservation(t *testing.T) {
	ctx := context.Background()
	alice, bob := newEngine(t), newEngine(t)
	first, err := bob.Bundle(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	retry, err := bob.Bundle(ctx, "alice")
	if err != nil || !bytes.Equal(first, retry) {
		t.Fatal("retry changed reserved bundle")
	}
	other, err := bob.Bundle(ctx, "charlie")
	if err != nil {
		t.Fatal(err)
	}
	var b1, b2 structpb.Struct
	if err := proto.Unmarshal(first, &b1); err != nil {
		t.Fatal(err)
	}
	if err := proto.Unmarshal(other, &b2); err != nil {
		t.Fatal(err)
	}
	if b1.Fields["opk_id"].GetNumberValue() == b2.Fields["opk_id"].GetNumberValue() {
		t.Fatal("OPK assigned to multiple peers")
	}

	var data structpb.Struct
	if err := proto.Unmarshal(first, &data); err != nil {
		t.Fatal(err)
	}
	data.Fields["spk"] = data.Fields["opk"]
	bad, err := proto.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	if err := alice.Establish(ctx, "bob", bad); err == nil {
		t.Fatal("invalid prekey signature accepted")
	}
	if alice.HasSession("bob") {
		t.Fatal("invalid bundle created a session")
	}
	if err := alice.Establish(ctx, "bob", first); err != nil {
		t.Fatal(err)
	}
}

func TestPadding(t *testing.T) {
	for _, size := range []int{0, 1, 252, 253, 1020, 1021, 4092, 4093, ratchet.MaxMessageSize} {
		message := bytes.Repeat([]byte("x"), size)
		padded, err := ratchet.Pad(message)
		if err != nil {
			t.Fatal(err)
		}
		plain, err := ratchet.Unpad(padded)
		if err != nil || !bytes.Equal(plain, message) {
			t.Fatalf("roundtrip size %d: %v", size, err)
		}
	}
	if _, err := ratchet.Pad(make([]byte, ratchet.MaxMessageSize+1)); err == nil {
		t.Fatal("oversized message accepted")
	}
	badLength := make([]byte, 256)
	binary.BigEndian.PutUint32(badLength, 253)
	badPadding := make([]byte, 256)
	badPadding[255] = 1
	for _, bad := range [][]byte{nil, {0}, make([]byte, 255), badLength, badPadding} {
		if _, err := ratchet.Unpad(bad); err == nil {
			t.Fatal("invalid padding accepted")
		}
	}
}

func TestMalformedEncryptedMessages(t *testing.T) {
	e := newEngine(t)
	for _, data := range [][]byte{nil, {protocol.PREKEY_TYPE}, {protocol.WHISPER_TYPE, 0}, {99, 0}, {protocol.PREKEY_TYPE, 0x33, 0x12, 0}} {
		if _, err := e.Decrypt(context.Background(), "remote", data); err == nil {
			t.Fatalf("accepted %x", data)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Bundle(ctx, "remote"); err != context.Canceled {
		t.Fatalf("canceled Bundle: %v", err)
	}
}

func FuzzDecrypt(f *testing.F) {
	e, err := ratchet.New()
	if err != nil {
		f.Fatal(err)
	}
	f.Add([]byte{protocol.WHISPER_TYPE, 0x33})
	f.Add([]byte{protocol.PREKEY_TYPE, 0x33, 0x12, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) <= 20*1024 {
			_, _ = e.Decrypt(context.Background(), "unknown", data)
		}
	})
}
