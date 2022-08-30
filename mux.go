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

	// GET /health
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
	// POST /login
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
	mux.Post("/login", l.ServeHTTP)

	// -- tasks --------------------------------
	// POST /tasks
	at := &handler.AddTask{
		Service:   &service.AddTask{DB: db, Repo: &r},
		Validator: v,
	}
	mux.Post("/tasks", at.ServeHTTP)

	// GET /tasks
	lt := &handler.ListTask{
		Service: &service.ListTask{DB: db, Repo: &r},
	}
	mux.Get("/tasks", lt.ServeHTTP)

	// -- users --------------------------------
	// POST /users
	ru := &handler.RegisterUser{
		Service:   &service.RegisterUser{DB: db, Repo: &r},
		Validator: v,
	}
	mux.Post("/users", ru.ServeHTTP)

	return mux, cleanup, nil
}
