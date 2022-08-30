package main

import (
	"context"
	"net/http"

	"github.com/ac0mz/go_todo_app/auth"
	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/config"
	"github.com/ac0mz/go_todo_app/handler"
	"github.com/ac0mz/go_todo_app/service"
	"github.com/ac0mz/go_todo_app/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func NewMux(ctx context.Context, cfg *config.Config) (http.Handler, func(), error) {
	mux := chi.NewRouter()

	// ヘルスチェックAPI
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// 静的解析エラー回避用に戻り値を明示的に破棄
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	db, cleanup, err := store.New(ctx, cfg)
	if err != nil {
		return nil, cleanup, err
	}
	clocker := clock.RealClocker{}
	r := store.Repository{Clocker: clocker}
	v := validator.New()

	// -- auth --------------------------------
	redisCli, err := store.NewKVS(ctx, cfg)
	if err != nil {
		return nil, cleanup, err
	}
	jwter, err := auth.NewJWTer(redisCli, clocker)
	if err != nil {
		return nil, cleanup, err
	}
	l := &handler.Login{
		Service:   &service.Login{DB: db, Repo: &r, TokenGenerator: jwter},
		Validator: v,
	}
	// 一般権限認証認可API
	mux.Post("/login", l.ServeHTTP)

	mux.Route("/admin", func(r chi.Router) {
		r.Use(handler.AuthMiddleware(jwter), handler.AdminMiddleware)
		// 管理者権限認証認可API
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			// 静的解析エラー回避用に戻り値を明示的に破棄
			_, _ = w.Write([]byte(`{"message": "admin only}"`))
		})
	})

	// -- tasks --------------------------------
	at := &handler.AddTask{
		Service:   &service.AddTask{DB: db, Repo: &r},
		Validator: v,
	}
	lt := &handler.ListTask{
		Service: &service.ListTask{DB: db, Repo: &r},
	}
	mux.Route("/tasks", func(r chi.Router) {
		// ログインしている場合のみ/tasksエンドポイントへのアクセスを許可する
		r.Use(handler.AuthMiddleware(jwter))
		// タスク個別登録API
		r.Post("/", at.ServeHTTP)
		// タスク一覧取得API
		r.Get("/", lt.ServeHTTP)
	})

	// -- users --------------------------------
	ru := &handler.RegisterUser{
		Service:   &service.RegisterUser{DB: db, Repo: &r},
		Validator: v,
	}
	// ユーザ個別登録API
	mux.Post("/users", ru.ServeHTTP)

	return mux, cleanup, nil
}
