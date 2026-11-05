package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

// TestWriteError_AppErrorReturnsCodeAndMessage کد و پیام کاربر AppError را برمی‌گرداند.
func TestWriteError_AppErrorReturnsCodeAndMessage(t *testing.T) {
	ginCtx := testutil.NewTestGinContext(http.MethodGet, "/api/example", nil)
	secret := "sql-internal-detail-9f3a"
	err := apperror.New("BAD_REQUEST", "درخواست نامعتبر است.", secret, http.StatusBadRequest)

	middleware.WriteError(ginCtx.Context, err)

	if ginCtx.Recorder.Code != http.StatusBadRequest {
		t.Fatalf("وضعیت HTTP باید ۴۰۰ باشد، دریافت %d", ginCtx.Recorder.Code)
	}
	code, msg := decodeErrorBody(t, ginCtx.Recorder.Body.Bytes())
	if code != "BAD_REQUEST" || msg != "درخواست نامعتبر است." {
		t.Fatalf("پاسخ خطا نادرست است: code=%s msg=%s", code, msg)
	}
	if bytes.Contains(ginCtx.Recorder.Body.Bytes(), []byte(secret)) {
		t.Fatal("جزئیات داخلی AppError در پاسخ آمده است")
	}
}

// TestWriteError_UnknownErrorIsGeneric500 خطای ناشناخته را ۵۰۰ و پیام عمومی می‌کند.
func TestWriteError_UnknownErrorIsGeneric500(t *testing.T) {
	ginCtx := testutil.NewTestGinContext(http.MethodPost, "/api/example", []byte(`{}`))
	secret := "dial failed password=SuperSecretInternal"
	middleware.WriteError(ginCtx.Context, errors.New(secret))

	if ginCtx.Recorder.Code != http.StatusInternalServerError {
		t.Fatalf("وضعیت HTTP باید ۵۰۰ باشد، دریافت %d", ginCtx.Recorder.Code)
	}
	code, msg := decodeErrorBody(t, ginCtx.Recorder.Body.Bytes())
	if code != apperror.ErrInternal.Code || msg != apperror.ErrInternal.UserMsg {
		t.Fatalf("پاسخ عمومی نادرست است: code=%s msg=%s", code, msg)
	}
	body := ginCtx.Recorder.Body.String()
	if bytes.Contains(ginCtx.Recorder.Body.Bytes(), []byte("SuperSecretInternal")) || bytes.Contains([]byte(body), []byte(secret)) {
		t.Fatal("متن خطای داخلی در پاسخ نشت کرده است")
	}
}

// decodeErrorBody فیلدهای error و code را از بدنه JSON می‌خواند.
func decodeErrorBody(t *testing.T, raw []byte) (code, msg string) {
	t.Helper()
	var body struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("بدنه JSON نیست: %s", raw)
	}
	return body.Code, body.Error
}
