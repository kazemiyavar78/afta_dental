package audit

import "testing"

type diffSample struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Note   string `json:"note"`
}

// TestDiff_OnlyChangedFields فیلد ثابت را حذف می‌کند و قبل و بعد فیلد عوض‌شده را نگه می‌دارد.
func TestDiff_OnlyChangedFields(t *testing.T) {
	got := Diff(
		diffSample{Name: "قدیم", Amount: 10, Note: "ثابت"},
		diffSample{Name: "جدید", Amount: 10, Note: "ثابت"},
		[]string{"name", "amount", "note"},
	)
	if len(got) != 1 {
		t.Fatalf("فقط نام باید فرق کند: %#v", got)
	}
	if got["name"].Before != "قدیم" || got["name"].After != "جدید" {
		t.Fatalf("تغییر نام نادرست است: %#v", got["name"])
	}
}

// TestDiff_NoChangeReturnsNil وقتی همه فیلدها یکی باشند رویدادی برای جزئیات نمی‌ماند.
func TestDiff_NoChangeReturnsNil(t *testing.T) {
	sample := diffSample{Name: "یک", Amount: 2}
	if Diff(sample, sample, []string{"name", "amount"}) != nil {
		t.Fatal("بدون تغییر باید nil باشد")
	}
	if Diff(sample, diffSample{Name: "دو"}, nil) != nil {
		t.Fatal("فهرست خالی باید nil باشد")
	}
}

// TestMaskTail_KeepsLastFourDigits چهار نویسه آخر می‌ماند و بقیه ستاره می‌شود.
func TestMaskTail_KeepsLastFourDigits(t *testing.T) {
	if got := MaskTail("09121234567", 4); got != "*******4567" {
		t.Fatalf("تلفن: %s", got)
	}
	if got := MaskTail("1234", 4); got != "1234" {
		t.Fatalf("متن کوتاه نباید عوض شود: %s", got)
	}
}
