package session

import (
	"testing"

	"github.com/google/uuid"
)

// TestTransposeUUIDEndianness تبدیل mixed-endian مایکروسافت به RFC 4122 را بررسی می‌کند.
func TestTransposeUUIDEndianness(t *testing.T) {
	// بایت‌های خام همان‌طور که درایور بدون تبدیل برمی‌گرداند
	// (معادل نمایش اشتباه ffb428f0-75c7-6b44-aba9-4b1cecc694ae)
	raw := [16]byte{
		0xff, 0xb4, 0x28, 0xf0,
		0x75, 0xc7,
		0x6b, 0x44,
		0xab, 0xa9, 0x4b, 0x1c, 0xec, 0xc6, 0x94, 0xae,
	}

	transposeUUIDEndianness(&raw)

	got := uuid.UUID(raw).String()
	want := "f028b4ff-c775-446b-aba9-4b1cecc694ae"
	if got != want {
		t.Fatalf("پس از transpose: got %s, want %s", got, want)
	}

	// برگشت‌پذیری: دوباره transpose باید به حالت قبل برگردد
	transposeUUIDEndianness(&raw)
	back := uuid.UUID(raw).String()
	wantBack := "ffb428f0-75c7-6b44-aba9-4b1cecc694ae"
	if back != wantBack {
		t.Fatalf("برگشت transpose: got %s, want %s", back, wantBack)
	}
}

// TestMSSQLUUIDScanFromBytes اسکن بایت‌های mixed-endian را بررسی می‌کند.
func TestMSSQLUUIDScanFromBytes(t *testing.T) {
	src := []byte{
		0xff, 0xb4, 0x28, 0xf0,
		0x75, 0xc7,
		0x6b, 0x44,
		0xab, 0xa9, 0x4b, 0x1c, 0xec, 0xc6, 0x94, 0xae,
	}

	var id MSSQLUUID
	if err := id.Scan(src); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	want := "f028b4ff-c775-446b-aba9-4b1cecc694ae"
	if id.String() != want {
		t.Fatalf("Scan result: got %s, want %s", id.String(), want)
	}
}

// TestMSSQLUUIDScanFromString اسکن رشته استاندارد (بدون جابه‌جایی) را بررسی می‌کند.
func TestMSSQLUUIDScanFromString(t *testing.T) {
	var id MSSQLUUID
	if err := id.Scan("F028B4FF-C775-446B-ABA9-4B1CECC694AE"); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	want := "f028b4ff-c775-446b-aba9-4b1cecc694ae"
	if id.String() != want {
		t.Fatalf("Scan string: got %s, want %s", id.String(), want)
	}
}

// TestMSSQLUUIDValueRoundTrip اطمینان می‌دهد Value بایت‌های mixed-endian می‌فرستد.
func TestMSSQLUUIDValueRoundTrip(t *testing.T) {
	id, err := ParseMSSQLUUID("f028b4ff-c775-446b-aba9-4b1cecc694ae")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	val, err := id.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}
	bytes, ok := val.([]byte)
	if !ok || len(bytes) != 16 {
		t.Fatalf("Value باید []byte با طول ۱۶ باشد، got %T", val)
	}

	want := []byte{
		0xff, 0xb4, 0x28, 0xf0,
		0x75, 0xc7,
		0x6b, 0x44,
		0xab, 0xa9, 0x4b, 0x1c, 0xec, 0xc6, 0x94, 0xae,
	}
	for i := range want {
		if bytes[i] != want[i] {
			t.Fatalf("Value[%d]: got %02x, want %02x (full=%x)", i, bytes[i], want[i], bytes)
		}
	}

	var scanned MSSQLUUID
	if err := scanned.Scan(bytes); err != nil {
		t.Fatalf("Scan round-trip: %v", err)
	}
	if scanned.String() != id.String() {
		t.Fatalf("round-trip: got %s, want %s", scanned.String(), id.String())
	}
}
