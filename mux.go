package main

import (
	"context"
	"net/http"

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
	r := store.Repository{Clocker: clock.RealClocker{}}
	v := validator.New()

	// -- tasks --------------------------------
	// POST /tasks
	at := &handler.AddTask{
		Service:   &service.AddTask{DB: db, Repo: &r},
		Validator: v,
	}
	mux.Post("/tasks", at.ServeHTTP)

	// GET /tasks
	lt := &handler.ListTask{
		// Service: &service.ListTask{DB: db, Repo: &r},
		Service: &service.ListTask{DB: db, Repo: &r},
	}
	mux.Get("/tasks", lt.ServeHTTP)

	// -- users --------------------------------
	// POST /tasks
	ru := &handler.RegisterUser{
		Service:   &service.RegisterUser{DB: db, Repo: &r},
		Validator: v,
	}
	mux.Post("/users", ru.ServeHTTP)

	return mux, cleanup, nil
}
