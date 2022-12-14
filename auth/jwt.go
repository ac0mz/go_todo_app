package auth

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
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

// GetToken はHTTPリクエストヘッダーのJWTを検証し、トークンを返却する
func (j JWTer) GetToken(ctx context.Context, r *http.Request) (jwt.Token, error) {
	// HTTPリクエストヘッダーからjwt.Tokenインターフェースを満たす型の値を取得
	token, err := jwt.ParseRequest(
		r,
		// 署名検証用アルゴリズムと公開鍵を指定
		jwt.WithKey(jwa.RS256, j.PublicKey),
		// 後続処理にて時刻情報(*auth.JWTer.Clocker)をベースに検証するため、ここでの検証は無視する
		jwt.WithValidate(false),
	)
	if err != nil {
		return nil, err
	}

	// 署名の検証
	if err = jwt.Validate(token, jwt.WithClock(j.Clocker)); err != nil {
		return nil, fmt.Errorf("GetToken: failed to validate token: %w", err)
	}
	// Redisに格納されたトークン存在有無チェック
	if _, err := j.Store.Load(ctx, token.JwtID()); err != nil {
		// 期限切れの他、手動で削除している場合もあり得る
		return nil, fmt.Errorf("GetToken: %q expired: %w", token.JwtID(), err)
	}
	return token, nil
}

type userIDKey struct{}
type roleKey struct{}

// SetUserID はcontext.Contextにキーバリューの形式でユーザIDを設定する
func SetUserID(ctx context.Context, uid entity.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, uid)
}

// GetUserID はcontext.ContextからユーザIDを取得し、値と取得成否を返却する
func GetUserID(ctx context.Context) (entity.UserID, bool) {
	// any型で取得した値をentity.UserIDで型アサーションし、かつ2つ目の戻り値を受け取ることでpanicを回避
	id, ok := ctx.Value(userIDKey{}).(entity.UserID)
	return id, ok
}

// SetRole はcontext.Contextにキーバリューの形式でロールを設定する
func SetRole(ctx context.Context, token jwt.Token) context.Context {
	role, ok := token.Get(RoleKey)
	if !ok {
		return context.WithValue(ctx, roleKey{}, "")
	}
	return context.WithValue(ctx, roleKey{}, role)
}

// GetRole はcontext.Contextからロールを取得し、値と取得成否を返却する
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey{}).(string)
	return role, ok
}

// FillContext は*http.Request型の値にユーザIDやロール権限の情報を設定する
func (j JWTer) FillContext(r *http.Request) (*http.Request, error) {
	token, err := j.GetToken(r.Context(), r)
	if err != nil {
		return nil, err
	}
	// リクエストヘッダーのJWTを元にRedisからユーザIDを取得
	uid, err := j.Store.Load(r.Context(), token.JwtID())
	if err != nil {
		return nil, err
	}
	ctx := SetUserID(r.Context(), uid)
	ctx = SetRole(ctx, token)

	// context.Context型の値を入れ替えた*http.Request型の値をディープコピー
	clone := r.Clone(ctx)
	return clone, nil
}

func IsAdmin(ctx context.Context) bool {
	role, ok := GetRole(ctx)
	if !ok {
		return false
	}
	return role == "admin"
}
