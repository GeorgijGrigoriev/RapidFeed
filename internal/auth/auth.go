package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/GeorgijGrigoriev/RapidFeed/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// maybe need to save feed news encrypted inside database, but now not used

var keyAES []byte

func Init() {
	slog.Info("initializing encryption subsystem")

	prepareAESKey(utils.SecretKey)

	if len(keyAES) == 0 {
		slog.Error("failed to initialize encryption subsystem, wrong key length")

		os.Exit(1)
	}

	slog.Info("encryption subsystem initialized successfully")
}

func Encrypt(password string) (string, error) {
	block, err := aes.NewCipher(keyAES)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(password), nil)

	encrypted := append(nonce, ciphertext...)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func Decrypt(password string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < 12 {
		return "", errors.New("not enough length")
	}

	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(keyAES)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func prepareAESKey(key string) {
	keyLen := len(key)

	switch keyLen {
	case 16:
		slog.Info("AES key length is 16 will use AES-128")

		keyAES = []byte(key)
	case 24:
		slog.Info("AES key length is 24 will use AES-192")

		keyAES = []byte(key)
	case 32:
		slog.Info("AES key length is 32 will use AES-256")

		keyAES = []byte(key)
	default:
		slog.Info("AES key length is invalid, please provide secret key with 16,24 or 32 length")

		keyAES = []byte("")
	}
}

func HashPassword(password string) ([]byte, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func CheckPassword(hash []byte, password string) error {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		slog.Error("failed to compare password", "error", err)

		return fmt.Errorf("incorrect password")
	}

	return nil
}
