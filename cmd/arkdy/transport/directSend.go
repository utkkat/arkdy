package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"
 
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	protocolID  = "/arkdy/direct/1.0.0"
	maxEnvelope = 20 * 1024
	ioTimeout   = 15 * time.Second
	dialTimeout = 30 * time.Second
)

func directSend(h host.Host, targetPeerID peer.ID, message string) error {
	targetAddr, err := ma.NewMultiaddr(string(targetPeerID))	
	if err != nil {
		return err
	}
	info, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		return err
	}
	h.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	stream, err := h.NewStream(context.Background(), info.ID, "/direct-send/1.0.0")
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
