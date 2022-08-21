package store

import (
	"errors"

	"github.com/ac0mz/go_todo_app/entity"
)

var (
	Tasks = &TaskStore{Tasks: map[entity.TaskID]*entity.Task{}}

	ErrNotFound = errors.New("not found")
)

type TaskStore struct {
	// 動作確認用のためあえてexport
	LastID entity.TaskID
	Tasks  map[entity.TaskID]*entity.Task
}

// Add はレコードを追加する
func (ts *TaskStore) Add(t *entity.Task) (entity.TaskID, error) {
	ts.LastID++
	t.ID = ts.LastID
	ts.Tasks[t.ID] = t
	return t.ID, nil
}

// All はソート済のタスク一覧を返す
func (ts *TaskStore) All() entity.Tasks {
	tasks := make([]*entity.Task, len(ts.Tasks))
	for i, t := range ts.Tasks {
		tasks[i-1] = t
	}
	return tasks
}
