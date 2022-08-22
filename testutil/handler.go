package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// AssertJSON は期待結果と実行結果のJSONボディにおける差分を比較検証する
func AssertJSON(t *testing.T, want, got []byte) {
	t.Helper()

	var wantJson, gotJson any
	if err := json.Unmarshal(want, &wantJson); err != nil {
		t.Fatalf("cannot unmarshal want %q: %v", want, err)
	}
	if err := json.Unmarshal(got, &gotJson); err != nil {
		t.Fatalf("cannot unmarshal got %q: %v", got, err)
	}
	if diff := cmp.Diff(wantJson, gotJson); diff != "" {
		t.Errorf("got differs: (-got +want)\n%s", diff)
	}
}

// AssertResponse はレスポンス情報を検証する
func AssertResponse(t *testing.T, got *http.Response, status int, body []byte) {
	t.Helper()
	t.Cleanup(func() { _ = got.Body.Close() })

	gotBody, err := io.ReadAll(got.Body)
	if err != nil {
		t.Fatal(err)
	}
	if got.StatusCode != status {
		t.Fatalf("want status %d, but got %d, body %q", status, got.StatusCode, gotBody)
	}

	if len(gotBody) == 0 && len(body) == 0 {
		return
	}
	AssertJSON(t, body, gotBody)
}

// LoadFile はファイルを読み込む
func LoadFile(t *testing.T, path string) []byte {
	t.Helper()

	byteData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read file %q: %v", path, err)
	}
	return byteData
}
