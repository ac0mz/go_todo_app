package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ac0mz/go_todo_app/entity"
	"github.com/go-playground/validator/v10"
)

type RegisterUser struct {
	Service   RegisterUserService
	Validator *validator.Validate
}

// ServeHTTP はハンドラー処理として、リクエストされたユーザ情報を登録し、登録結果のIDをレスポンス情報として作成する
func (ru RegisterUser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// リクエストボディのデシリアライズ
	var b struct {
		Name     string `json:"name" validate:"required"`
		Password string `json:"password" validate:"required"`
		Role     string `json:"role" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		RespondJSON(ctx, w, &ErrResponse{Message: err.Error()}, http.StatusInternalServerError)
		return
	}

	// DB登録
	u, err := ru.Service.RegisterUser(ctx, b.Name, b.Password, b.Role)
	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	// レスポンス
	rsp := struct {
		ID entity.UserId
	}{ID: u.ID}
	RespondJSON(ctx, w, rsp, http.StatusOK)
}
