package auth

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/store"
	"github.com/ac0mz/go_todo_app/testutil/fixture"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
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

func TestJWTer_GetToken(t *testing.T) {
	t.Parallel()

	c := clock.FixedClocker{}
	// 期待結果のトークンとテスト入力値の署名付きトークン
	want, signed := createToken(t, c)
	// モック設定
	moq := &StoreMock{}
	moq.LoadFunc = func(ctx context.Context, key string) (entity.UserID, error) {
		return entity.UserID(20), nil
	}
	sut, err := NewJWTer(moq, c)
	if err != nil {
		t.Fatal(err)
	}
	// テスト入力値の設定
	req := createRequest(signed)

	// 実行と検証
	got, err := sut.GetToken(context.Background(), req)
	if err != nil {
		t.Fatalf("want no error, but got %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetToken() got = %v, want = %v", got, want)
	}
}

type FixedTomorrowClocker struct{}

func (c FixedTomorrowClocker) Now() time.Time {
	return clock.FixedClocker{}.Now().Add(24 * time.Hour)
}

// TestJWTer_GetToken_NG トークンが有効期限切れ、およびRedis上に存在しない場合のケース
func TestJWTer_GetToken_NG(t *testing.T) {
	t.Parallel()

	c := clock.FixedClocker{}
	// テスト入力値の署名付きトークン
	_, signed := createToken(t, c)
	// テスト入力値の設定
	req := createRequest(signed)

	// テストデータ作成
	type moq struct {
		userID entity.UserID
		err    error
	}
	tests := map[string]struct {
		c   clock.Clocker
		moq moq
	}{
		"expire": {
			// トークンのexpirationより未来時刻を返す
			c: FixedTomorrowClocker{},
		},
		"notFoundInStore": {
			c:   clock.FixedClocker{},
			moq: moq{err: store.ErrNotFound},
		},
	}

	for n, tt := range tests {
		tt := tt
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			// モック設定
			moq := &StoreMock{}
			moq.LoadFunc = func(ctx context.Context, key string) (entity.UserID, error) {
				return tt.moq.userID, tt.moq.err
			}
			sut, err := NewJWTer(moq, tt.c)
			if err != nil {
				t.Fatal(err)
			}

			// 実行と検証
			got, err := sut.GetToken(context.Background(), req)
			if err == nil {
				t.Errorf("want error, but got nil")
			}
			if got != nil {
				t.Errorf("want nil, but got %v", got)
			}
		})
	}
}

// createToken はテストデータのトークンを生成する
func createToken(t *testing.T, c clock.Clocker) (jwt.Token, []byte) {
	token, err := jwt.NewBuilder().
		JwtID(uuid.New().String()).
		Issuer(`github.com/ac0mz/go_todo_app`).
		Subject("access_token").
		IssuedAt(c.Now()).
		Expiration(c.Now().Add(30*time.Minute)).
		Claim(RoleKey, "test").
		Claim(UserNameKey, "test_user").
		Build()
	if err != nil {
		t.Fatal(err)
	}
	// 署名付きトークンの生成
	pkey, err := jwk.ParseKey(rawPrivKey, jwk.WithPEM(true))
	if err != nil {
		t.Fatal(err)
	}
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, pkey))
	if err != nil {
		t.Fatal(err)
	}
	return token, signed
}

// createRequest は署名付きトークンがヘッダーに設定されたリクエスト情報を生成する
func createRequest(signed []byte) *http.Request {
	req := httptest.NewRequest(
		http.MethodGet,
		`https://github.com/ac0mz/go_todo_app`,
		nil,
	)
	req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, signed))
	return req
}
