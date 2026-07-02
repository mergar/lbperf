package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"regexp"
)

var (
	keyPattern   = regexp.MustCompile(`^[a-zA-Z0-9=+/]+$`)
	tokenPattern = regexp.MustCompile(`^[0-9a-fA-F]+$`)
)

func ComputeToken(key, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

func ValidateToken(key, token, secret string) bool {
	expected := ComputeToken(key, secret)
	if len(expected) != len(token) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(token)) == 1
}

func ParseKey(raw string) (string, bool) {
	if len(raw) < 10 || len(raw) > 64 {
		return "", false
	}
	if !keyPattern.MatchString(raw) {
		return "", false
	}
	return raw, true
}

func ParseToken(raw string) (string, bool) {
	if len(raw) < 10 || len(raw) > 64 {
		return "", false
	}
	if !tokenPattern.MatchString(raw) {
		return "", false
	}
	return raw, true
}
