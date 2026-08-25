package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

func HashPassword(
	password string,
) (string, error) {
	if password == "" {
		return "", errors.New("密码不能为空")
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	digest := digest(password, salt)
	return fmt.Sprintf("sha256$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(digest)), nil
}

func CheckPassword(
	encoded,
	password string,
) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 3 || parts[0] != "sha256" {
		return false
	}
	salt, errSalt :=
		base64.RawStdEncoding.DecodeString(
			parts[1])
	want, errWant :=
		base64.RawStdEncoding.DecodeString(
			parts[2])
	if errSalt != nil || errWant != nil {
		return false
	}
	got := digest(password, salt)
	if len(got) != len(want) {
		return false
	}
	var same byte
	for index := range got {
		same |= got[index] ^ want[index]
	}
	return same == 0
}

func digest(
	password string,
	salt []byte,
) []byte {
	value := append(append([]byte{}, salt...), []byte(password)...)
	for index := 0; index < 12000; index++ {
		hash := sha256.Sum256(value)
		value = append(hash[:],
			salt...)
	}
	return value[:32]
}
