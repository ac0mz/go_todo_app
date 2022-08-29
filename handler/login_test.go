package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ac0mz/go_todo_app/testutil"
	"github.com/go-playground/validator/v10"
)

func TestLogin_ServeHTTP(t *testing.T) {
	type moq struct {
		token string
		err   error
	}
	type want struct {
		status  int
		rspFile string
	}

	tests := map[string]struct {
		reqFile string
		moq     moq
		want    want
	}{
		"ok": {
			reqFile: "testdata/login/status200_req.json.golden",
			moq: moq{
				token: "from_moq",
			},
			want: want{
				status:  http.StatusOK,
				rspFile: "testdata/login/status200_rsp.json.golden",
			},
		},
		"badRequest": {
			reqFile: "testdata/login/status400_req.json.golden",
			want: want{
				status:  http.StatusBadRequest,
				rspFile: "testdata/login/status400_rsp.json.golden",
			},
		},
		"internalServerError": {
			reqFile: "testdata/login/status200_req.json.golden",
			moq: moq{
				err: errors.New("error from mock"),
			},
			want: want{
				status:  http.StatusInternalServerError,
				rspFile: "testdata/login/status500_rsp.json.golden",
			},
		},
	}
	for n, tt := range tests {
		tt := tt
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			// モック設定
			moq := &LoginServiceMock{}
			moq.LoginFunc = func(ctx context.Context, name string, password string) (string, error) {
				return tt.moq.token, tt.moq.err
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				"/login",
				bytes.NewReader(testutil.LoadFile(t, tt.reqFile)),
			)

			sut := Login{
				Service:   moq,
				Validator: validator.New(),
			}
			// 実行と検証
			sut.ServeHTTP(w, r)
			rsp := w.Result()
			testutil.AssertResponse(t, rsp, tt.want.status, testutil.LoadFile(t, tt.want.rspFile))
		})
	}
}
