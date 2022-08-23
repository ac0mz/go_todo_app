package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

func TestRepository_ListTasks(t *testing.T) {
	ctx := context.Background()

	// entity.Taskを作成する他ケースの実行タイミングと重複してDB操作結果が変わる可能性があるため、
	// トランザクションを張ることで、当該テストケースの中だけのテーブル状態にする
	tx, err := testutil.OpenDBForTest(t).BeginTxx(ctx, nil)
	// 当該ケース完了後にDB状態を元に戻す
	t.Cleanup(func() { _ = tx.Rollback() })
	if err != nil {
		t.Fatal(err)
	}

	// DB状態の初期化と期待結果の準備
	wants := prepareTask(ctx, t, tx)

	// 実行
	sut := &Repository{}
	gots, err := sut.ListTasks(ctx, tx)
	if err != nil {
		t.Fatalf("unexecuted error: %v", err)
	}

	// 検証
	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s", d)
	}
}

// prepareTask はDBテストデータの仕込みを実行する
func prepareTask(ctx context.Context, t *testing.T, con Execer) entity.Tasks {
	// t.Helper()

	// DB状態を初期化
	if _, err := con.ExecContext(ctx, "DELETE FROM tasks;"); err != nil {
		t.Logf("failed initialize tasks: %v", err)
	}
	// テストデータ準備
	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{Title: "want task 1", Status: "todo", Created: c.Now(), Modified: c.Now()},
		{Title: "want task 2", Status: "todo", Created: c.Now(), Modified: c.Now()},
		{Title: "want task 3", Status: "done", Created: c.Now(), Modified: c.Now()},
	}
	// DB登録
	// ※INSERT文末尾のセミコロンを忘れるだけでpanicが発生するため要注意
	result, err := con.ExecContext(ctx,
		`INSERT INTO tasks (title, status, created, modified)
				VALUES (?, ?, ?, ?), (?, ?, ?, ?), (?, ?, ?, ?);`,
		wants[0].Title, wants[0].Status, wants[0].Created, wants[0].Modified,
		wants[1].Title, wants[1].Status, wants[1].Created, wants[1].Modified,
		wants[2].Title, wants[2].Status, wants[2].Created, wants[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	// MySQLでは複数レコード挿入時にLastInsertId()で取得される値は1件目のid値となる
	// wants[0].ID = entity.TaskID(id)
	// wants[1].ID = entity.TaskID(id + 1)
	// wants[2].ID = entity.TaskID(id + 2)
	for i, want := range wants {
		want.ID = entity.TaskID(id + int64(i))
	}
	return wants
}

func TestRepository_AddTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// データ準備
	c := clock.FixedClocker{}
	var wantID int64 = 20
	okTask := &entity.Task{
		Title: "ok task", Status: "todo", Created: c.Now(), Modified: c.Now(),
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// モック設定
	mock.ExpectExec(
		// DATA-DOG/go-sqlmock の仕様上、エスケープが必要
		`INSERT INTO tasks \(title, status, created, modified\) VALUES \(\?, \?, \?, \?\)`,
	).
		WithArgs(okTask.Title, okTask.Status, okTask.Created, okTask.Modified).
		WillReturnResult(sqlmock.NewResult(wantID, 1))

	xdb := sqlx.NewDb(db, "mysql")
	// 固定された時刻情報を使用してタスク登録を実行する
	r := &Repository{Clocker: c}
	if err := r.AddTask(ctx, xdb, okTask); err != nil {
		t.Errorf("want no error, but got %v", err)
	}

	// 	検証
	if okTask.ID != entity.TaskID(wantID) {
		t.Errorf("differs: (-got +want)\n%d", okTask.ID)
	}
}
