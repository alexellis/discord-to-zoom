package function

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	sdk "github.com/openfaas/go-sdk"
)

// verify verifies the request signature from Discord, this method will be
// called periodically by Discord to verify the endpoint is still active.
// Discord will also send invalid requests to test the endpoint is functioning
// correctly.
func verify(w http.ResponseWriter, r *http.Request, body []byte) {

	// https://discord.com/developers/docs/interactions/receiving-and-responding#security-and-authorization

	publicKey, err := sdk.ReadSecret("discord-public-key")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	signatureHexDecoded, err := hex.DecodeString(signature)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	if len(signatureHexDecoded) != ed25519.SignatureSize {
		http.Error(w, "invalid signature length", http.StatusUnauthorized)
		log.Println("invalid signature length")
		return
	}

	publicKeyHexDecoded, err := hex.DecodeString(publicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Println(err)
		return
	}

	log.Println(len(publicKeyHexDecoded))

	pubKey := [32]byte{}

	copy(pubKey[:], []byte(publicKeyHexDecoded))

	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(body)

	verified := ed25519.Verify(publicKeyHexDecoded, msg.Bytes(), signatureHexDecoded)
	log.Printf("Ping verified? %v\n", verified)

	if !verified {
		http.Error(w, "invalid request signature", http.StatusUnauthorized)
		return
	}

	p := map[string]float64{
		"type": float64(1),
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(p); err != nil {
		log.Print(err)
	}

}
