package db

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func (repo *JobRepository) encrypt(secret string) (string, string, error) {
	errs := func(err error) (string, string, error) {
		return "", "", err
	}

	block, err := repo.createBlock()
	if err != nil {
		return errs(err)
	}

	cipherText := make([]byte, aes.BlockSize+len(secret))
	_, err = io.ReadFull(rand.Reader, cipherText)
	if err != nil {
		return errs(err)
	}
	iv := cipherText[:aes.BlockSize]
	ivHex := hex.EncodeToString(iv)

	encrypter := cipher.NewCFBEncrypter(block, iv)

	encrypted := make([]byte, len(secret))
	encrypter.XORKeyStream(encrypted, []byte(secret))

	encryptedBytes := hex.EncodeToString(encrypted)
	return string(encryptedBytes), ivHex, nil
}

func (repo *JobRepository) decrypt(encrypted, ivHex string) (string, error) {
	encryptedBytes, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	iv, err := hex.DecodeString(ivHex)
	if err != nil {
		return "", err
	}

	block, err := repo.createBlock()
	if err != nil {
		return "", err
	}

	decrypter := cipher.NewCFBDecrypter(block, iv)

	decrypted := make([]byte, 1024)
	decrypter.XORKeyStream(decrypted, encryptedBytes)

	secret := string(bytes.Trim(decrypted, "\x00"))
	return secret, nil
}

func hash(password string) []byte {
	var hashedPassword []byte
	result := sha256.Sum256([]byte(password))
	hashedPassword = result[:]
	return hashedPassword
}

func (repo *JobRepository) createBlock() (cipher.Block, error) {
	return aes.NewCipher([]byte(repo.SkeletonKey))
}
