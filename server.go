package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
	l   net.Listener
}

func NewServer(l net.Listener, mux http.Handler) *Server {
	return &Server{
		srv: &http.Server{Handler: mux},
		l:   l,
	}
}

// Run はHTTPサーバを起動する関数
func (s *Server) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)
	// 別ゴルーチンでHTTPサーバを起動する
	eg.Go(func() error {
		// http.ErrServerClosed は s.srv.Shutdown(context.Background()) の正常終了を示している
		if err := s.srv.Serve(s.l); err != nil && err != http.ErrServerClosed {
			log.Printf("filed to closed: %+v", err)
			return err
		}
		return nil
	})

	// チャネルからの終了通知を待機する
	<-ctx.Done()
	if err := s.srv.Shutdown(context.Background()); err != nil {
		log.Printf("failed to shutdown: %+v", err)
	}
	// Goメソッドで起動した別ゴルーチンの終了を待機する(グレースフルシャットダウン)
	return eg.Wait()
}
