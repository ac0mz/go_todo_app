package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Login struct {
	Service   LoginService
	Validator *validator.Validate
}

func (l *Login) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var body struct {
		UserName string `json:"user_name" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondJSON(ctx, w, &ErrResponse{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	// 必須チェック
	if err := l.Validator.Struct(body); err != nil {
		RespondJSON(ctx, w, &ErrResponse{Message: err.Error()}, http.StatusBadRequest)
		return
	}

	// ログイン
	jwt, err := l.Service.Login(ctx, body.UserName, body.Password)
	if err != nil {
		RespondJSON(ctx, w, &ErrResponse{Message: err.Error()}, http.StatusInternalServerError)
		return
	}

	rsp := struct {
		AccessToken string `json:"access_token"`
	}{
		AccessToken: jwt,
	}
	RespondJSON(ctx, w, &rsp, http.StatusOK)
}
