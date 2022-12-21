package main

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

func ReadPeerKey(peerKey string) (crypto.PrivKey, error) {
	bytes, err := crypto.ConfigDecodeKey(peerKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(bytes)
}

func GeneratePeerKey() (string, string, error) {
	privk, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	if err != nil {
		return "", "", err
	}

	bytes, err := crypto.MarshalPrivateKey(privk)
	if err != nil {
		return "", "", err
	}

	id, _ := peer.IDFromPrivateKey(privk)

	return crypto.ConfigEncodeKey(bytes), id.String(), nil
}
