package provider

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"log"
)

func parsePrivateKey(clientPrivateKey string) (*rsa.PrivateKey, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(clientPrivateKey)

	if err != nil {
		return nil, errors.New("failed to decode base64 encoded private key: " + err.Error())
	}

	keyBytes, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)

	if err != nil {
		return nil, errors.New("failed to parse PKCS#8 private key: " + err.Error())
	}

	privateKey, ok := keyBytes.(*rsa.PrivateKey)
	if !ok {
		log.Fatal("not an RSA private key")
	}

	return privateKey, nil
}

func parsePublicKey(serverPublicKey string) (*rsa.PublicKey, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(serverPublicKey)

	if err != nil {
		return nil, errors.New("failed to decode base64 encoded public key: " + err.Error())
	}

	keyBytes, err := x509.ParsePKIXPublicKey(publicKeyBytes)

	if err != nil {
		return nil, errors.New("failed to parse PKCS#8 public key: " + err.Error())
	}

	publicKey, ok := keyBytes.(*rsa.PublicKey)
	if !ok {
		log.Fatal("not an RSA private key")
	}

	return publicKey, nil
}
