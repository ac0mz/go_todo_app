package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ac0mz/go_todo_app/clock"
	"github.com/ac0mz/go_todo_app/entity"
	"github.com/ac0mz/go_todo_app/testutil"
	"github.com/ac0mz/go_todo_app/testutil/fixture"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

func TestRepository_ListTasks(t *testing.T) {
	t.Parallel()

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
	wantUserID, wants := prepareTask(ctx, t, tx)

	// 実行
	sut := &Repository{}
	gots, err := sut.ListTasks(ctx, tx, wantUserID)
	if err != nil {
		t.Fatalf("unexecuted error: %v", err)
	}

	// 検証
	if d := cmp.Diff(gots, wants); len(d) != 0 {
		t.Errorf("differs: (-got +want)\n%s", d)
	}
}

// prepareUser はDBテストデータの仕込みを実行する
func prepareUser(ctx context.Context, t *testing.T, db Execer) entity.UserID {
	t.Helper()
	u := fixture.User(nil)
	result, err := db.ExecContext(ctx,
		`INSERT INTO users (name, password, role, created, modified)
		VALUES (?, ?, ?, ?, ?);`, u.Name, u.Password, u.Role, u.Created, u.Modified,
	)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("got user_id: %v", err)
	}
	return entity.UserID(id)
}

// prepareTask はDBテストデータの仕込みを実行する
func prepareTask(ctx context.Context, t *testing.T, db Execer) (entity.UserID, entity.Tasks) {
	t.Helper()
	// // DB状態を初期化
	// if _, err := db.ExecContext(ctx, "DELETE FROM tasks;"); err != nil {
	// 	t.Logf("failed initialize tasks: %v", err)
	// }
	userID := prepareUser(ctx, t, db)
	otherUserID := prepareUser(ctx, t, db)
	c := clock.FixedClocker{}
	wants := entity.Tasks{
		{
			UserID: userID, Title: "want task 1", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		}, {
			UserID: userID, Title: "want task 2", Status: "done",
			Created: c.Now(), Modified: c.Now(),
		},
	}
	tasks := entity.Tasks{
		wants[0],
		{
			UserID: otherUserID, Title: "not want task", Status: "todo",
			Created: c.Now(), Modified: c.Now(),
		},
		wants[1],
	}
	// DB登録
	// ※INSERT文末尾のセミコロンを忘れるだけでpanicが発生するため要注意
	result, err := db.ExecContext(ctx,
		`INSERT INTO tasks (user_id, title, status, created, modified)
				VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?), (?, ?, ?, ?, ?);`,
		tasks[0].UserID, tasks[0].Title, tasks[0].Status, tasks[0].Created, tasks[0].Modified,
		tasks[1].UserID, tasks[1].Title, tasks[1].Status, tasks[1].Created, tasks[1].Modified,
		tasks[2].UserID, tasks[2].Title, tasks[2].Status, tasks[2].Created, tasks[2].Modified,
	)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	// MySQLでは複数レコード挿入時にLastInsertId()で取得される値は1件目のid値となる
	// 期待結果として反映させるため、tasks経由で[]wants.IDにLastInsertIdを格納
	for i, task := range tasks {
		task.ID = entity.TaskID(id + int64(i))
	}
	return userID, wants
}

func TestRepository_AddTask(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// データ準備
	c := clock.FixedClocker{}
	var wantID int64 = 20
	okTask := &entity.Task{
		UserID: 3, Title: "ok task", Status: "todo", Created: c.Now(), Modified: c.Now(),
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// モック設定
	mock.ExpectExec(
		// DATA-DOG/go-sqlmock の仕様上、エスケープが必要
		`INSERT INTO tasks \(user_id, title, status, created, modified\)
				VALUES \(\?, \?, \?, \?, \?\)`,
	).
		WithArgs(okTask.UserID, okTask.Title, okTask.Status, okTask.Created, okTask.Modified).
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
