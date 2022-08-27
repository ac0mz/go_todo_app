package service

import (
	"context"

	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/store"
)

// 以下インターフェースはstore/task.goに実装する

//go:generate go run github.com/matryer/moq -out moq_test.go . TaskLister TaskAdder UserRegister
type TaskLister interface {
	ListTasks(ctx context.Context, db store.Queryer) (entity.Tasks, error)
}

type TaskAdder interface {
	AddTask(ctx context.Context, db store.Execer, t *entity.Task) error
}

type UserRegister interface {
	RegisterUser(ctx context.Context, db store.Execer, u *entity.User) error
}
