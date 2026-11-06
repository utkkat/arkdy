package relay

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func GetPrivateKey(path string) (crypto.PrivKey, error) {
	if path == "" {
		return nil, errors.New("relay private key path is empty")
	}

	key, err := LoadPrivateKey(path)
	if err == nil {
		return key, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("load relay private key: %w", err)
	}

	key, _, err = crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate relay private key: %w", err)
	}
	if err := savePrivateKey(path, key); err != nil {
		if errors.Is(err, os.ErrExist) {
			return LoadPrivateKey(path)
		}
		return nil, fmt.Errorf("save relay private key: %w", err)
	}
	return key, nil
}

func savePrivateKey(path string, key crypto.PrivKey) error {
	data, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(path), ".arkdy-relay-key-*")
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	return os.Link(file.Name(), path)
}

func LoadPrivateKey(path string) (crypto.PrivKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPrivateKey(data)
}
