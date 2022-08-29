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
	getUser = `SELECT id, name, password, role, created, modified FROM users WHERE name = ?`
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
	u.ID = entity.UserID(id)
	return nil
}

func (r *Repository) GetUser(ctx context.Context, db Queryer, name string) (*entity.User, error) {
	u := &entity.User{}
	if err := db.GetContext(ctx, u, getUser, name); err != nil {
		return nil, err
	}
	return u, nil
}
