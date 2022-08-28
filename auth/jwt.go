package auth

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

//go:embed cert/secret.pem
var rawPrivKey []byte

//go:embed cert/public.pem
var rawPubKey []byte

// JWTer はPEMキーから変換したJWKと、KVSに保存するStoreインターフェースを持つ
type JWTer struct {
	PrivateKey, PublicKey jwk.Key
	Store                 Store
	Clocker               clock.Clocker
}

//go:generate go run github.com/matryer/moq -out moq_test.go . Store
type Store interface {
	Save(ctx context.Context, key string, userID entity.UserID) error
	Load(ctx context.Context, key string) (entity.UserID, error)
}

// NewJWTer はPEMキーの解析と構造体の初期化を行う
func NewJWTer(s Store, c clock.Clocker) (*JWTer, error) {
	privKey, err := parse(rawPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed in NewJWTer: private key: %w", err)
	}
	pubKey, err := parse(rawPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed in NewJWTer: public key: %w", err)
	}

	j := &JWTer{
		Store:      s,
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Clocker:    c,
	}
	return j, nil
}

// parse はPEMキーを解析し、JWKに変換する
func parse(rawKey []byte) (jwk.Key, error) {
	key, err := jwk.ParseKey(rawKey, jwk.WithPEM(true))
	if err != nil {
		return nil, err
	}
	return key, nil
}

const (
	RoleKey     = "role"
	UserNameKey = "user_name"
)

// GenerateToken はユーザ情報と秘密鍵を元にJWTトークンを生成する。
// また、トークン生成時に作成したUUID（JWT ID）をキーにRedisへユーザIDを登録する。
func (j JWTer) GenerateToken(ctx context.Context, u entity.User) ([]byte, error) {
	token, err := jwt.NewBuilder().
		JwtID(uuid.New().String()).
		Issuer(`github.com/ac0mz/go_todo_app`).
		Subject("access_token").
		IssuedAt(j.Clocker.Now()).
		Expiration(j.Clocker.Now().Add(30*time.Minute)).
		Claim(RoleKey, u.Role).     // 独自クレーム(ロール)
		Claim(UserNameKey, u.Name). // 独自クレーム(ユーザ名)
		Build()
	if err != nil {
		return nil, fmt.Errorf("GetToken: failed to build token: %w", err)
	}

	// UUIDをキーにユーザIDを登録
	if err := j.Store.Save(ctx, token.JwtID(), u.ID); err != nil {
		return nil, err
	}

	// 秘密鍵による署名を付与したJWTトークンの生成
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, j.PrivateKey))
	if err != nil {
		return nil, err
	}
	return signed, nil
}
