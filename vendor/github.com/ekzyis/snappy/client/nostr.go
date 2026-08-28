package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

type nostrEvent struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// decodeNsec decodes a NIP-19 nsec (bech32) into its 32-byte private key.
func decodeNsec(nsec string) ([]byte, error) {
	hrp, data, err := bech32Decode(nsec)
	if err != nil {
		return nil, fmt.Errorf("invalid nsec: %w", err)
	}
	if hrp != "nsec" {
		return nil, fmt.Errorf("expected nsec, got %s", hrp)
	}
	key, err := bech32ConvertBits(data, 5, 8, false)
	if err != nil {
		return nil, fmt.Errorf("invalid nsec: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("expected 32-byte key, got %d bytes", len(key))
	}
	return key, nil
}

func signAuthEvent(nsec, k1, u string) (string, error) {
	sk, err := decodeNsec(nsec)
	if err != nil {
		return "", err
	}

	privKey, pubKey := btcec.PrivKeyFromBytes(sk)

	event := nostrEvent{
		PubKey:    hex.EncodeToString(schnorr.SerializePubKey(pubKey)),
		CreatedAt: time.Now().Unix(),
		Kind:      27235,
		Tags: [][]string{
			{"challenge", k1},
			{"u", u},
			{"method", "GET"},
		},
		Content: "Stacker News Authentication",
	}

	id, err := event.hash()
	if err != nil {
		return "", err
	}
	event.ID = hex.EncodeToString(id)

	sig, err := schnorr.Sign(privKey, id)
	if err != nil {
		return "", fmt.Errorf("error signing auth event: %w", err)
	}
	event.Sig = hex.EncodeToString(sig.Serialize())

	signed, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("error encoding auth event: %w", err)
	}
	return string(signed), nil
}

func (e nostrEvent) hash() ([]byte, error) {
	var buf strings.Builder
	enc := json.NewEncoder(&buf)
	// nostr's canonical form does not HTML-escape < > &
	enc.SetEscapeHTML(false)
	if err := enc.Encode([]interface{}{0, e.PubKey, e.CreatedAt, e.Kind, e.Tags, e.Content}); err != nil {
		return nil, fmt.Errorf("error serializing auth event: %w", err)
	}
	sum := sha256.Sum256([]byte(strings.TrimRight(buf.String(), "\n")))
	return sum[:], nil
}
