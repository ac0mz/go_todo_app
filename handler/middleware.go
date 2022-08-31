package handler

import (
	"net/http"

	"github.com/ac0mz/go_todo_app/auth"
)

// AuthMiddleware はcontext.Context型の値にユーザ情報を埋め込むミドルウェア
func AuthMiddleware(j *auth.JWTer) func(next http.Handler) http.Handler {
	// クロージャでシグネチャを合わせた関数を返す
	return func(next http.Handler) http.Handler {
		// アクセストークンが見つからなかった場合、当該関数でリクエスト処理を終了するため認証も兼ねている
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ユーザIDおよびロール権限をcontext.Context型の値に設定した*http.Request型の値を取得
			req, err := j.FillContext(r)
			if err != nil {
				RespondJSON(r.Context(), w, ErrResponse{
					Message: "not find auth info",
					Details: []string{err.Error()},
				}, http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}

// AdminMiddleware はロール権限を確認するミドルウェア
// context.Context型の値にユーザ情報が埋め込まれていることが前提で呼び出される想定
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAdmin(r.Context()) {
			// 管理者権限ではない場合
			RespondJSON(r.Context(), w, ErrResponse{
				Message: "not admin",
			}, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
