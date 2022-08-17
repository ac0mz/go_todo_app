package main

import (
	context "context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"golang.org/x/sync/errgroup"
)

func Test_run_OK(t *testing.T) {
	// 前処理
	cancel, eg, l := doRun(t)

	// エンドポイントにGETリクエストを発行
	in := "message"
	rsp := sendGetRequest(t, l, in)
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
	cancel, eg, l := doRun(t)

	in := "message"
	rsp := sendGetRequest(t, l, in)
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
func doRun(t *testing.T) (context.CancelFunc, *errgroup.Group, net.Listener) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to listen: %+v", err)
	}
	// キャンセル可能なcontext.Contextを生成
	ctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでrun関数を実行し、HTTPサーバを起動
	eg.Go(func() error {
		return run(ctx, l)
	})
	return cancel, eg, l
}

// sendGetRequest はエンドポイントにGETリクエストを発行する
func sendGetRequest(t *testing.T, l net.Listener, in string) *http.Response {
	// URL生成
	url := fmt.Sprintf("http://%s/%s", l.Addr().String(), in)
	t.Logf("try request to %q", url)
	// リクエスト発行
	rsp, err := http.Get(url)
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
