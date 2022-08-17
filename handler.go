package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
)

// run はHTTPサーバを起動する関数
func run(ctx context.Context, l net.Listener) error {
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
		}),
	}
	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでHTTPサーバを起動する
	eg.Go(func() error {
		// http.ErrServerClosed は s.Shutdown(context.Background()) の正常終了を示している
		if err := s.Serve(l); err != nil && err != http.ErrServerClosed {
			log.Printf("filed to closed: %v", err)
			return err
		}
		return nil
	})

	// チャネルからの終了通知を待機する
	<-ctx.Done()
	if err := s.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}
	// Goメソッドで起動した別ゴルーチンの終了を待機する
	return eg.Wait()
}

func main() {
	// 入力バリデーション
	if 1 < len(os.Args) {
		log.Printf("need port number\n")
		os.Exit(1)
	}
	// 動的にポート番号を取得してrun関数を起動
	port := os.Args[1]
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("failed to listen port %s: %v", port, err)
	}
	if err := run(context.Background(), l); err != nil {
		log.Printf("failed to terminate server: %v", err)
	}
}
