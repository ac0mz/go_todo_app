package auth

import (
	"bytes"
	"testing"
)

func TestEmbed(t *testing.T) {
	// 公開鍵
	wantBegin := []byte("-----BEGIN PUBLIC KEY-----")
	if !bytes.Contains(rawPubKey, wantBegin) {
		t.Errorf("want %s, but got %s", wantBegin, rawPubKey)
	}
	wantEnd := []byte("-----END PUBLIC KEY-----")
	if !bytes.Contains(rawPubKey, wantEnd) {
		t.Errorf("want %s, but got %s", wantEnd, rawPubKey)
	}

	// 秘密鍵
	wantBegin = []byte("-----BEGIN PRIVATE KEY-----")
	if !bytes.Contains(rawPrivKey, wantBegin) {
		t.Errorf("want %s, but got %s", wantBegin, rawPrivKey)
	}
	wantEnd = []byte("-----END PRIVATE KEY-----")
	if !bytes.Contains(rawPrivKey, wantEnd) {
		t.Errorf("want %s, but got %s", wantEnd, rawPrivKey)
	}
}
