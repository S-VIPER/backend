package verification

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"

	"github.com/S-VIPER/backend/gin-api/internal/service"
)

type CodeGenerator struct{}

func NewCodeGenerator() *CodeGenerator {
	return &CodeGenerator{}
}

var _ service.VerificationCodeGenerator = (*CodeGenerator)(nil)

func (g *CodeGenerator) Generate() (string, string, error) {
	var buf [4]byte

	if _, err := rand.Read(buf[:]); err != nil {
		return "", "", fmt.Errorf("generate verification code: %w", err)
	}

	number :=
		uint32(buf[0])<<24 |
			uint32(buf[1])<<16 |
			uint32(buf[2])<<8 |
			uint32(buf[3])
	number %= 1_000_000

	plainCode := fmt.Sprintf("%06d", number)

	hash := sha256.Sum256([]byte(plainCode))
	codeHash := hex.EncodeToString(hash[:])

	return plainCode, codeHash, nil
}

func (g *CodeGenerator) Compare(
	codeHash string,
	code string,
) bool {
	hash := sha256.Sum256([]byte(code))
	actualHash := hex.EncodeToString(hash[:])

	return subtle.ConstantTimeCompare(
		[]byte(codeHash),
		[]byte(actualHash),
	) == 1
}
