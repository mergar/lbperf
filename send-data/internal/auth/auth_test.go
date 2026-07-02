package auth

import "testing"

func TestComputeTokenDeterministic(t *testing.T) {
	secret := "_SuperMegaGigaPuperMyB_MegaSecret!"
	key := "SDFrDgl0U6UEpTJuw9zURwQFTSe8bHMplY4fPVxkdU="
	token := ComputeToken(key, secret)
	if len(token) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(token))
	}
	if !ValidateToken(key, token, secret) {
		t.Fatal("token should validate")
	}
}

func TestParseTokenRejectsNonHex(t *testing.T) {
	if _, ok := ParseToken("zzzzzzzzzz"); ok {
		t.Fatal("expected reject")
	}
}
