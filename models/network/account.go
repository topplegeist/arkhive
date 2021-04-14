package network

import (
	"crypto/rsa"
	"time"
)

type Account struct {
	Username         string
	Email            string
	RegistrationDate time.Time
	PrivateKey       rsa.PrivateKey
	PublicKey        rsa.PublicKey
	Sign             []byte
}
