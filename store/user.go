package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/ac0mz/go_todo_app/entity"
	"github.com/go-sql-driver/mysql"
)

const (
	insertUser = `INSERT INTO users (name, password, role, created, modified)
			 VALUES (?, ?, ?, ?, ?);`
)

func (r *Repository) RegisterUser(ctx context.Context, db Execer, u *entity.User) error {
	u.Created = r.Clocker.Now()
	u.Modified = r.Clocker.Now()

	result, err := db.ExecContext(ctx, insertUser,
		u.Name, u.Password, u.Role, u.Created, u.Modified)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == ErrCodeMySQLDuplicateEntry {
			return fmt.Errorf("cannot create same name user: %w", ErrAlreadyEntry)
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = entity.UserId(id)
	return nil
}
