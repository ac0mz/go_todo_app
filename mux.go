package main

import (
	"net/http"

	"github.com/ac0mz/go_todo_app/handler"
	"github.com/ac0mz/go_todo_app/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func NewMux() http.Handler {
	mux := chi.NewRouter()

	// GET: /health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// 静的解析エラー回避用に戻り値を明示的に破棄
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})

	// POST: /tasks
	v := validator.New()
	at := &handler.AddTask{Store: store.Tasks, Validator: v}
	mux.Post("/tasks", at.ServeHTTP)

	// GET: /tasks
	lt := &handler.ListTask{Store: store.Tasks}
	mux.Get("/tasks", lt.ServeHTTP)

	return mux
}
