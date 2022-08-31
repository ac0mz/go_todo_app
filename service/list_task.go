package service

import (
	"context"
	"fmt"

	"github.com/ac0mz/go_todo_app/auth"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/store"
)

type ListTask struct {
	DB   store.Queryer
	Repo TaskLister
}

// ListTasks は一意のユーザに紐付いたタスク一覧のみを取得する
// handler/service.goの実装
func (l *ListTask) ListTasks(ctx context.Context) (entity.Tasks, error) {
	id, ok := auth.GetUserID(ctx)
	if !ok {
		return nil, fmt.Errorf("user_id not found")
	}
	ts, err := l.Repo.ListTasks(ctx, l.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %v", err)
	}
	return ts, nil
}
