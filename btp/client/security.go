package client

import (
	"bytes"
	"context"
	"crypto/rsa"
	"io"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// EncryptPayload encrypt an arbitrary message to a JWE-encrypted payload
// #ref https://help.sap.com/docs/credential-store/sap-credential-store/code-samples-go
func EncryptPayload(ctx context.Context, body any, publicKey rsa.PublicKey) (*bytes.Buffer, error) {

	opts := new(jose.EncrypterOptions)
	opts.WithHeader("iat", time.Now().Unix())

	encrypter, err := jose.NewEncrypter(jose.A256GCM, jose.Recipient{Algorithm: jose.RSA_OAEP_256, Key: &publicKey}, opts)

	if err != nil {
		tflog.Error(ctx, "failed to instantiate encrypter:", map[string]any{
			"algorithm": jose.A256GCM,
			"key_type":  jose.RSA_OAEP_256,
		})
		return nil, err
	}

	payload, err := encodeJson(body)

	if err != nil {
		tflog.Error(ctx, "failed to encode payload")
		return nil, err
	}

	jwe, err := encrypter.Encrypt([]byte(*payload))

	if err != nil {
		tflog.Error(ctx, "failed to encrypt payload")
		return nil, err
	}

	jweCompact, err := jwe.CompactSerialize()

	if err != nil {
		tflog.Error(ctx, "failed to serialize payload as compact")
		return nil, err
	}

	return bytes.NewBufferString(jweCompact), nil
}

// DecryptResponse decrypt the JWE response to a buffer
// #ref https://help.sap.com/docs/credential-store/sap-credential-store/code-samples-go
func DecryptResponse(ctx context.Context, body io.Reader, privateKey rsa.PrivateKey) (io.Reader, error) {
	ciphertext := new(strings.Builder)
	_, err := io.Copy(ciphertext, body)

	if err != nil {
		return nil, err
	}

	object, err := jose.ParseEncryptedCompact(ciphertext.String(), []jose.KeyAlgorithm{jose.RSA_OAEP_256}, []jose.ContentEncryption{jose.A256GCM})

	if err != nil {
		tflog.Error(ctx, "failed to parse encrypted metadata:", map[string]any{
			"algorithm": jose.A256GCM,
			"key_type":  jose.RSA_OAEP_256,
		})
		return nil, err
	}

	decrypted, err := object.Decrypt(&privateKey)

	if err != nil {
		tflog.Error(ctx, "failed to decrypt payload")
		return nil, err
	}

	return bytes.NewReader(decrypted), nil
}
