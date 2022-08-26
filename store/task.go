package store

import (
	"context"

	"github.com/ac0mz/go_todo_app/entity"
)

const (
	selectAllTasks = `SELECT id, title, status, created, modified FROM tasks;`
	insertTask     = `INSERT INTO tasks (title, status, created, modified) VALUES (?, ?, ?, ?);`
)

// 以下はservice/interface.goの実装

// ListTasks は*entity.Task型の値をすべて取得し、スライスで返却する
func (r *Repository) ListTasks(ctx context.Context, db Queryer) (entity.Tasks, error) {
	tasks := entity.Tasks{}
	if err := db.SelectContext(ctx, &tasks, selectAllTasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// AddTask は1件のタスクを登録し、引数で渡された*entity.Task.IDに発行されたIDを格納する
func (r *Repository) AddTask(ctx context.Context, db Execer, t *entity.Task) error {
	t.Created = r.Clocker.Now()
	t.Modified = r.Clocker.Now()
	result, err := db.ExecContext(ctx, insertTask, t.Title, t.Status, t.Created, t.Modified)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	t.ID = entity.TaskID(id)
	return nil
}
