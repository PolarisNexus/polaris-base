package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
)

func verifyRS256(key *rsa.PublicKey, signed, sig []byte) error {
	h := sha256.Sum256(signed)
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, h[:], sig)
}
