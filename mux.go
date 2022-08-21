package main

import "net/http"

func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		// 静的解析エラー回避用に戻り値を明示的に破棄
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	return mux
}
