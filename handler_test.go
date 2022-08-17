package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_run_OK(t *testing.T) {
	// 前処理
	cancel, eg := doRun()

	// エンドポイントにGETリクエストを発行
	in := "message"
	rsp := sendGetRequest(t, in)
	defer rsp.Body.Close()

	// レスポンス取得
	got, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %+v", err)
	}
	// 検証: 文字列が期待結果と一致すること
	want := fmt.Sprintf("Hello, %s!", in)
	if string(got) != want {
		t.Errorf("want %q, but got %q", want, got)
	}

	// 後処理
	assertRun(t, cancel, eg)
}

func Test_run_NG(t *testing.T) {
	// 前処理
	cancel, eg := doRun()

	in := "message"
	rsp := sendGetRequest(t, in)
	defer rsp.Body.Close()

	// レスポンス取得
	got, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read body: %+v", err)
	}
	// 検証: 文字列が期待結果と不一致であること
	want := "Hello, World!"
	if string(got) == want {
		t.Errorf("want %q, but got %q", want, got)
	}

	// 後処理
	assertRun(t, cancel, eg)
}

// doRun はrun関数を実行する
func doRun() (context.CancelFunc, *errgroup.Group) {
	// キャンセル可能なcontext.Contextを生成
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでrun関数を実行し、HTTPサーバを起動
	eg.Go(func() error {
		return run(ctx)
	})
	return cancel, eg
}

// sendGetRequest はエンドポイントにGETリクエストを発行する
func sendGetRequest(t *testing.T, in string) *http.Response {
	rsp, err := http.Get("http://localhost:18080/" + in)
	if err != nil {
		t.Errorf("failed to get: %+v", err)
	}
	return rsp
}

// assertRun はrun関数の実行を終了し、結果を検証する
func assertRun(t *testing.T, cancel context.CancelFunc, eg *errgroup.Group) {
	// cancel()による終了通知送信
	cancel()
	// サーバを停止し、run関数の戻り値を検証
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
}
