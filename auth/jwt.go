package auth

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/lestrrat-go/jwx/v2/jwk"
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
func NewJWTer(s Store) (*JWTer, error) {
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
		Clocker:    clock.RealClocker{},
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
