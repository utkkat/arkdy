package main

import (
	"crypto/rand"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

type EphKey struct {
	priv crypto.PrivKey
	ID   peer.ID
}

func GenerateEphemeralKey() (*EphKey, error) {
	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil,err
	}
	peerID, err := peer.IDFromPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return &EphKey{
		priv: priv,
		ID:   peerID,
	}, nil
}
