package bsphp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func MD5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}

func AES128CBCEncryptBase64(plaintext, key16 string) (string, error) {
	if len(key16) < 16 {
		return "", errors.New("key too short")
	}
	key := []byte(key16[:16])
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	padded := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	iv := key
	mode := cipher.NewCBCEncrypter(block, iv)
	out := make([]byte, len(padded))
	mode.CryptBlocks(out, padded)
	return base64.StdEncoding.EncodeToString(out), nil
}

func AES128CBCDecryptBase64ToString(ciphertextB64, key16 string) (string, error) {
	if len(key16) < 16 {
		return "", errors.New("key too short")
	}
	ct, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	key := []byte(key16[:16])
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	if len(ct)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext not multiple of block")
	}
	iv := key
	mode := cipher.NewCBCDecrypter(block, iv)
	out := make([]byte, len(ct))
	mode.CryptBlocks(out, ct)
	out, err = pkcs7Unpad(out, aes.BlockSize)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func RSAEncryptPKCS1Base64(message, publicKeyBase64DER string) (string, error) {
	pub, err := parseRSAPublicKey(publicKeyBase64DER)
	if err != nil {
		return "", err
	}
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(message))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func RSADecryptPKCS1Base64(ciphertextB64, privateKeyBase64DER string) (string, error) {
	priv, err := parseRSAPrivateKey(privateKeyBase64DER)
	if err != nil {
		return "", err
	}
	ct, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	out, err := rsa.DecryptPKCS1v15(rand.Reader, priv, ct)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	if pad == 0 {
		pad = blockSize
	}
	p := make([]byte, pad)
	for i := range p {
		p[i] = byte(pad)
	}
	return append(data, p...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty")
	}
	pad := int(data[len(data)-1])
	if pad > blockSize || pad == 0 {
		return nil, errors.New("bad padding")
	}
	for i := len(data) - pad; i < len(data); i++ {
		if int(data[i]) != pad {
			return nil, errors.New("bad padding bytes")
		}
	}
	return data[:len(data)-pad], nil
}

func cleanKeyB64(s string) string {
	s = strings.ReplaceAll(s, "-----BEGIN RSA PRIVATE KEY-----", "")
	s = strings.ReplaceAll(s, "-----BEGIN PRIVATE KEY-----", "")
	s = strings.ReplaceAll(s, "-----BEGIN PUBLIC KEY-----", "")
	s = strings.ReplaceAll(s, "-----END RSA PRIVATE KEY-----", "")
	s = strings.ReplaceAll(s, "-----END PRIVATE KEY-----", "")
	s = strings.ReplaceAll(s, "-----END PUBLIC KEY-----", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}

func parseRSAPublicKey(b64 string) (*rsa.PublicKey, error) {
	raw, err := base64.StdEncoding.DecodeString(cleanKeyB64(b64))
	if err != nil {
		return nil, err
	}
	k, err := x509.ParsePKIXPublicKey(raw)
	if err != nil {
		return nil, err
	}
	pub, ok := k.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not rsa public key")
	}
	return pub, nil
}

func parseRSAPrivateKey(b64 string) (*rsa.PrivateKey, error) {
	raw, err := base64.StdEncoding.DecodeString(cleanKeyB64(b64))
	if err != nil {
		return nil, err
	}
	if k, err := x509.ParsePKCS8PrivateKey(raw); err == nil {
		if priv, ok := k.(*rsa.PrivateKey); ok {
			return priv, nil
		}
	}
	return x509.ParsePKCS1PrivateKey(raw)
}
