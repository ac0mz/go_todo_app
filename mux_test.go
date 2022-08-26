package main

import (
	"context"
	"github.com/ac0mz/go_todo_app/config"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_NewMux(t *testing.T) {
	ctx := context.Background()
	// HTTPサーバを実際に起動せずにテストするためのモック生成
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)

	cfg, err := config.New()
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}
	// ハンドラのルータ（コントローラ）であるmuxを生成
	mux, _, err := NewMux(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create mux: %v", err)
	}

	// リクエスト送信とレスポンス受信
	mux.ServeHTTP(w, r)                        // モックを引数に、mux生成時のHandleFuncを呼び出し
	res := w.Result()                          // ハンドラで生成されたレスポンスを返却
	t.Cleanup(func() { _ = res.Body.Close() }) // deferのように全処理が完了後、登録された関数を実行
	if res.StatusCode != http.StatusOK {
		t.Error("want status code 200, but", res.StatusCode)
	}

	// 検証
	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v", err)
	}

	want := `{"status": "ok"}`
	if string(got) != want {
		t.Errorf("want %q, but got %q", want, got)
	}
}
