package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/testutil"
)

func TestKVS_Save(t *testing.T) {
	t.Parallel()

	// データ準備
	key := "TestKVS_Save"
	uid := entity.UserID(1234)
	ctx := context.Background()
	cli := testutil.OpenRedisForTest(t)
	t.Cleanup(func() {
		cli.Del(ctx, key)
	})
	// テスト対象の準備
	sut := &KVS{Cli: cli}

	// 検証
	if err := sut.Save(ctx, key, uid); err != nil {
		t.Errorf("want no error, but got %v", err)
	}
}

func TestKVS_Load(t *testing.T) {
	t.Parallel()

	cli := testutil.OpenRedisForTest(t)
	sut := &KVS{Cli: cli}

	t.Run("ok", func(t *testing.T) {
		t.Parallel()

		// データ準備
		key := "TestKVS_Load_ok"
		uid := entity.UserID(1234)
		ctx := context.Background()
		cli.Set(ctx, key, int64(uid), 30*time.Minute)
		t.Cleanup(func() {
			cli.Del(ctx, key)
		})

		// 実行と検証
		got, err := sut.Load(ctx, key)
		if err != nil {
			t.Fatalf("want no error, but got %v", err)
		}
		if got != uid {
			t.Errorf("want: %d, but got: %d", uid, got)
		}
	})

	t.Run("notFound", func(t *testing.T) {
		t.Parallel()

		// データ準備
		key := "TestKVS_Load_notFound"
		ctx := context.Background()

		// 実行と検証
		got, err := sut.Load(ctx, key)
		if err == nil || !errors.Is(err, ErrNotFound) {
			t.Errorf("want: %v, but got: %v(value = %d)", ErrNotFound, err, got)
		}
	})
}
