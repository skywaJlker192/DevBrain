package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher struct {
	memory      uint32
	iterations  uint32
	saltLength  uint8
	keyLength   uint32
}

// ТРЕБОВАНИЕ 3: Хеширование через argon2
func NewPasswordHasher(memory uint32, iterations uint32, saltLen uint8, keyLen uint32) *PasswordHasher {
	return &PasswordHasher{
		memory:      memory,
		iterations:  iterations,
		saltLength:  saltLen,
		keyLength:   keyLen,
	}
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	// Генерация случайной соли
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Argon2id - лучшая защита
	hash := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, 4, h.keyLength)

	// Формат: $argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHRzYWx0$aGVsbG93b3JsZA
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=4$%s$%s",
		h.memory, h.iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *PasswordHasher) Compare(password, hashString string) (bool, error) {
	parts := strings.Split(hashString, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	// Извлечение параметров
	var version, memory, iterations uint32
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, err
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=4", &memory, &iterations); err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}

	// Генерация хеша для сравнения
	comparisonHash := argon2.IDKey([]byte(password), salt, iterations, memory, 4, uint32(len(decodedHash)))

	// ТРЕБОВАНИЕ 3: Constant-time comparison
	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}
