package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/ac0mz/go_todo_app/config"
)

// run はHTTPサーバを起動する関数
func run(ctx context.Context) error {
	// 環境変数の読み込み
	cfg, err := config.New()
	if err != nil {
		return err
	}
	// HTTP通信を待機
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed to listen port %d: %v", cfg.Port, err)
	}
	url := fmt.Sprintf("http://%s", l.Addr().String())
	log.Printf("start with: %v", url)

	// handlerをルーティングするmuxの生成
	mux, cleanup, err := NewMux(ctx, cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	// HTTPサーバの生成と起動
	s := NewServer(l, mux)
	return s.Run(ctx)
}

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("failed to terminate server: %v", err)
		os.Exit(1)
	}
}
