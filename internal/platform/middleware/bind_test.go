package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

type d08Payload struct {
	Name string `json:"name" form:"name" binding:"required"`
	Age  int    `json:"age" form:"age"`
}

// captureLog خروجی استاندارد لاگ را در بافر می‌گیرد تا مقدار فیلد در آن بررسی شود.
// t تست جاری است. خروجی بافر لاگ همین بازه است و در پایان به نویسنده قبلی برمی‌گردد.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(prev) })
	return &buf
}

// assertValidationError وضعیت ۴۰۰، کد VALIDATION_ERROR و نبود جزئیات فیلد را چک می‌کند.
// body بدنه پاسخ است. secret اگر تهی نباشد نباید در پاسخ یا لاگ باشد. logs خروجی لاگ سرور است.
func assertValidationError(t *testing.T, status int, body []byte, logs string, secret string) {
	t.Helper()
	if status != http.StatusBadRequest {
		t.Fatalf("وضعیت HTTP باید ۴۰۰ باشد، دریافت %d", status)
	}
	code, msg := decodeErrorBody(t, body)
	if code != apperror.ErrValidation.Code || msg != apperror.ErrValidation.UserMsg {
		t.Fatalf("پاسخ اعتبارسنجی نادرست است: code=%s msg=%s", code, msg)
	}
	text := string(body)
	if strings.Contains(text, "Name") || strings.Contains(text, "name") || strings.Contains(text, "required") {
		t.Fatalf("جزئیات فیلد در پاسخ آمده است: %s", text)
	}
	if secret != "" && (strings.Contains(text, secret) || strings.Contains(logs, secret)) {
		t.Fatalf("مقدار فیلد در پاسخ یا لاگ آمده است")
	}
}

// TestD08_BindJSON_EmptyOrIncompleteReturnsValidationError بدنه خالی و فیلد required ناقص را ۴۰۰ می‌کند.
func TestD08_BindJSON_EmptyOrIncompleteReturnsValidationError(t *testing.T) {
	logs := captureLog(t)

	t.Run("empty body", func(t *testing.T) {
		logs.Reset()
		ginCtx := testutil.NewTestGinContext(http.MethodPost, "/api/example", nil)
		var got d08Payload
		if middleware.BindJSON(ginCtx.Context, &got) {
			t.Fatal("بدنه خالی باید false برگرداند")
		}
		assertValidationError(t, ginCtx.Recorder.Code, ginCtx.Recorder.Body.Bytes(), logs.String(), "")
	})

	t.Run("missing required", func(t *testing.T) {
		logs.Reset()
		ginCtx := testutil.NewTestGinContext(http.MethodPost, "/api/example", []byte(`{}`))
		var got d08Payload
		if middleware.BindJSON(ginCtx.Context, &got) {
			t.Fatal("فیلد required ناقص باید false برگرداند")
		}
		assertValidationError(t, ginCtx.Recorder.Code, ginCtx.Recorder.Body.Bytes(), logs.String(), "")
		if !strings.Contains(logs.String(), "Name") || !strings.Contains(logs.String(), "required") {
			t.Fatalf("لاگ سرور باید نام فیلد و قاعده را داشته باشد: %s", logs.String())
		}
	})
}

// TestD08_BindJSON_ValidBodyFillsStruct بدنه معتبر را می‌پذیرد و struct را پر می‌کند.
func TestD08_BindJSON_ValidBodyFillsStruct(t *testing.T) {
	ginCtx := testutil.NewTestGinContext(http.MethodPost, "/api/example", []byte(`{"name":"نرمین","age":4}`))
	var got d08Payload
	if !middleware.BindJSON(ginCtx.Context, &got) {
		t.Fatal("بدنه معتبر باید true برگرداند")
	}
	if got.Name != "نرمین" || got.Age != 4 {
		t.Fatalf("struct پر نشده است: %+v", got)
	}
	if ginCtx.Recorder.Body.Len() != 0 {
		t.Fatalf("بدنه معتبر نباید پاسخ خطا داشته باشد: %s", ginCtx.Recorder.Body.String())
	}
}

// TestD08_BindQuery_InvalidValueDoesNotLeakFieldValue مقدار query نامعتبر را به کاربر و لاگ نمی‌دهد.
func TestD08_BindQuery_InvalidValueDoesNotLeakFieldValue(t *testing.T) {
	const secret = "secret-bind-value-9f3a"
	logs := captureLog(t)
	ginCtx := testutil.NewTestGinContext(http.MethodGet, "/api/example?age="+secret, nil)
	var got d08Payload
	if middleware.BindQuery(ginCtx.Context, &got) {
		t.Fatal("query نامعتبر باید false برگرداند")
	}
	assertValidationError(t, ginCtx.Recorder.Code, ginCtx.Recorder.Body.Bytes(), logs.String(), secret)
}
