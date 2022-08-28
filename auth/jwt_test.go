package auth

import (
	"bytes"
	"context"
	"testing"

	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/testutil/fixture"
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

func TestJWTer_GenerateToken(t *testing.T) {
	ctx := context.Background()
	wantID := entity.UserID(20)
	u := fixture.User(&entity.User{ID: wantID})

	moq := &StoreMock{}
	moq.SaveFunc = func(ctx context.Context, key string, userID entity.UserID) error {
		if userID != wantID {
			t.Errorf("want %d, but got %d", wantID, userID)
		}
		return nil
	}

	sut, err := NewJWTer(moq, clock.FixedClocker{})
	if err != nil {
		t.Fatal(err)
	}

	// 実行と検証
	got, err := sut.GenerateToken(ctx, *u)
	if err != nil {
		t.Errorf("not want err: %v", err)
	}
	if len(got) == 0 {
		t.Errorf("got token is empty")
	}
}
