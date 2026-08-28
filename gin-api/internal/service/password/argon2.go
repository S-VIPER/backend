package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/S-VIPER/backend/gin-api/internal/service"
	"golang.org/x/crypto/argon2"
)

var ErrInvalidPassword = errors.New("invalid password")

type Argon2Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
	saltLength  uint32
}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		keyLength:   32,
		saltLength:  16,
	}
}

var _ service.PasswordHasher = (*Argon2Hasher)(nil)

func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.iterations,
		h.memory,
		h.parallelism,
		h.keyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.memory,
		h.iterations,
		h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func (h *Argon2Hasher) Compare(
	encodedHash string,
	password string,
) error {
	// $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return fmt.Errorf("invalid argon2 hash format")
	}

	if parts[1] != "argon2id" {
		return fmt.Errorf("unsupported password hash algorithm: %s", parts[1])
	}

	if parts[2] != "v=19" {
		return fmt.Errorf("unsupported argon2 version: %s", parts[2])
	}

	params := strings.Split(parts[3], ",")
	if len(params) != 3 {
		return fmt.Errorf("invalid argon2 parameters")
	}

	memory, err := parseUint32Param(params[0], "m")
	if err != nil {
		return err
	}

	iterations, err := parseUint32Param(params[1], "t")
	if err != nil {
		return err
	}

	parallelism, err := parseUint8Param(params[2], "p")
	if err != nil {
		return err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("decode argon2 salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return fmt.Errorf("decode argon2 hash: %w", err)
	}

	if len(salt) == 0 {
		return fmt.Errorf("argon2 salt is empty")
	}

	if len(expectedHash) == 0 {
		return fmt.Errorf("argon2 hash is empty")
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(actualHash, expectedHash) != 1 {
		return ErrInvalidPassword
	}

	return nil
}

func parseUint32Param(value string, name string) (uint32, error) {
	prefix := name + "="

	if !strings.HasPrefix(value, prefix) {
		return 0, fmt.Errorf("invalid argon2 parameter: %s", value)
	}

	result, err := strconv.ParseUint(
		strings.TrimPrefix(value, prefix),
		10,
		32,
	)
	if err != nil {
		return 0, fmt.Errorf("invalid argon2 parameter %s: %w", name, err)
	}

	if result == 0 {
		return 0, fmt.Errorf("argon2 parameter %s must be greater than zero", name)
	}

	return uint32(result), nil
}

func parseUint8Param(value string, name string) (uint8, error) {
	prefix := name + "="

	if !strings.HasPrefix(value, prefix) {
		return 0, fmt.Errorf("invalid argon2 parameter: %s", value)
	}

	result, err := strconv.ParseUint(
		strings.TrimPrefix(value, prefix),
		10,
		8,
	)
	if err != nil {
		return 0, fmt.Errorf("invalid argon2 parameter %s: %w", name, err)
	}

	if result == 0 {
		return 0, fmt.Errorf("argon2 parameter %s must be greater than zero", name)
	}

	return uint8(result), nil
}

var _ interface {
	Hash(string) (string, error)
	Compare(string, string) error
} = (*Argon2Hasher)(nil)
