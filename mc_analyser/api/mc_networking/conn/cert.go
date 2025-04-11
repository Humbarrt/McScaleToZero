package conn

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"mc_analyser/api/mc_networking/conn/crypto"
	"mc_analyser/logging"
	"strings"
)

type certify struct {
	privateKey      *rsa.PrivateKey
	privateKeyBytes []byte
	publicKeyBytes  []byte
	verifyToken     []byte
	isUsed          bool
	sharedSecret    []byte

	encrypt cipher.Stream
	decrypt cipher.Stream
}

func newCert() (*certify, error) {
	_cert := &certify{}
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logging.Error("Error generating RSA key:", "error", err)
		return _cert, err
	}
	privateKey.Precompute()
	if err := privateKey.Validate(); err != nil {
		logging.Error("Error validating RSA key:", "error", err)
		return _cert, err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		logging.Error("Error marshalling public key:", "error", err)
		return _cert, err
	}

	verifyToken := make([]byte, 4)
	_, err = rand.Read(verifyToken)
	if err != nil {
		logging.Error("Error generating verify token:", "error", err)
		return _cert, err
	}

	logging.Info("Generated verify token", "token", verifyToken, "length", len(verifyToken))

	return &certify{
		privateKey:      privateKey,
		privateKeyBytes: privateKeyBytes,
		publicKeyBytes:  publicKeyBytes,
		verifyToken:     verifyToken,
		isUsed:          false,
	}, nil
}

func (cert *certify) IsUsed() bool {
	return cert.isUsed
}

func (c *certify) GetPublicKey() []byte {
	if c.publicKeyBytes == nil {
		logging.Error("Public key is not generated")
		return nil
	}
	return c.publicKeyBytes
}

func (c *certify) GetPrivateKey() []byte {
	if c.privateKeyBytes == nil {
		logging.Error("Private key is not generated")
		return nil
	}
	return c.privateKeyBytes
}

func (c *certify) GetVerifyToken() []byte {
	if c.verifyToken == nil {
		logging.Error("Verify token is not generated")
		return nil
	}
	return c.verifyToken
}

func (c *certify) GetServerId() (string, error) {
	if !c.isUsed {
		return "", fmt.Errorf("authentication is not used")
	}
	sha := sha1.New()
	sha.Write(c.sharedSecret)
	sha.Write(c.publicKeyBytes)
	hash := sha.Sum(nil)

	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80

	if negative {
		carry := true

		for i := len(hash) - 1; i >= 0; i-- {
			hash[i] = ^hash[i]
			if carry {
				carry = hash[i] == 0xff
				hash[i]++
			}
		}
	}

	// Trim away zeroes
	res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
	if negative {
		res = "-" + res
	}

	return res, nil
}

func (c *certify) AuthEncrypt(data []byte) ([]byte, error) {
	if c.privateKey == nil {
		return nil, fmt.Errorf("private key is not generated")
	}
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, &c.privateKey.PublicKey, data)
	if err != nil {
		logging.Error("Error encrypting data:", "error", err)
		return nil, err
	}
	return encrypted, nil
}

func (c *certify) AuthDecrypt(encrypted []byte) ([]byte, error) {
	if c.privateKey == nil {
		return nil, fmt.Errorf("private key is not generated")
	}
	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, encrypted)
	if err != nil {
		logging.Error("Error decrypting data:", "error", err)
		return nil, err
	}
	return decrypted, nil
}

func (c *certify) EnableEncryption(sharedSecret []byte) error {
	c.sharedSecret = sharedSecret
	encrypt, decrypt, err := crypto.NewEncryptAndDecrypt(sharedSecret)
	if err != nil {
		logging.Error("Error creating encrypt/decrypt:", "error", err)
		return err
	}

	c.encrypt = encrypt
	c.decrypt = decrypt
	c.isUsed = true
	return err
}

func (c *certify) Encrypt(data []byte) (output []byte) {
	if !c.isUsed {
		return data
	}

	output = make([]byte, len(data))
	c.encrypt.XORKeyStream(output, data)
	return
}

func (c *certify) Decrypt(data []byte) (output []byte) {
	if !c.isUsed {
		return data
	}

	output = make([]byte, len(data))
	c.decrypt.XORKeyStream(output, data)
	return
}
