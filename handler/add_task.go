package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/store"
	"github.com/go-playground/validator/v10"
)

type AddTask struct {
	Store     *store.TaskStore
	Validator *validator.Validate
}

func (at *AddTask) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var b struct {
		Title string `json:"title" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		// リクエスト情報の解析が失敗した場合
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	err := at.Validator.Struct(b)
	if err != nil {
		// バリデーションエラーの場合
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusBadRequest)
		return
	}

	t := &entity.Task{
		Title:   b.Title,
		Status:  entity.TaskStatusTodo,
		Created: time.Now(),
	}
	id, err := store.Tasks.Add(t)
	if err != nil {
		// DB操作が失敗した場合
		RespondJSON(ctx, w, &ErrResponse{
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}
	rsp := struct {
		ID entity.TaskID `json:"id"`
	}{ID: id}
	RespondJSON(ctx, w, rsp, http.StatusOK)
}
