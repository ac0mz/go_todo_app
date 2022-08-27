package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	driverName    = "mysql"
	dataSourceFmt = "%s:%s@tcp(%s:%d)/%s?parseTime=true"

	// ErrCodeMySQLDuplicateEntry はMySQLにおけるDUPLICATEエラーコード
	// https://dev.mysql.com/doc/mysql-erros/8.0/en/server-error-reference.html
	// Error number: 1062; Symbol: ER_DUP_ENTRY; SQLSTATE: 23000
	ErrCodeMySQLDuplicateEntry = 1062
)

var (
	ErrAlreadyEntry = errors.New("duplicate entry")
)

func New(ctx context.Context, cfg *config.Config) (*sqlx.DB, func(), error) {
	db, err := sql.Open(
		driverName,
		fmt.Sprintf(dataSourceFmt, cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName),
	)
	if err != nil {
		return nil, nil, err
	}

	// Openは接続テストが実行されないため、db.PingContextで明示的に疎通確認を実行する
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, func() { _ = db.Close() }, err
	}

	xdb := sqlx.NewDb(db, driverName)
	return xdb, func() { _ = db.Close() }, nil
}

// Beginner はトランザクションの開始操作を扱う
type Beginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// Preparer is an interface used by Preparex.
// prepared statementとしてSQLを扱う(標準パッケージのsql.Stmtをラップする)
type Preparer interface {
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
}

// Execer is an interface used by MustExec and LoadFile.
// 書き込み系の操作を扱う
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// Queryer is an interface used by Get and Select
// 参照系の操作を扱う
// dest interface{} は取得結果の格納先となる構造体のポインタ型を指定する
type Queryer interface {
	Preparer
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...any) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...any) error
}

var (
	// interfaceが期待通りに宣言されていることの検証用コード
	// *sqlx.DB型をnilで初期化した値を右辺で作成後、各interfaceに代入することでコンパイラに検証させる
	// 以下の書き方で作成する場合は、ポインタ型の値を作成する方法と異なりメモリアロケーションが発生しない
	_ Beginner = (*sqlx.DB)(nil)
	_ Preparer = (*sqlx.DB)(nil)
	_ Execer   = (*sqlx.DB)(nil)
	_ Queryer  = (*sqlx.DB)(nil)
	_ Queryer  = (*sqlx.Tx)(nil)
)

// Repository はすべてのDB操作を扱う
type Repository struct {
	Clocker clock.Clocker // SQL実行時の時刻情報を制御する
}
