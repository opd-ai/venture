package engine

import (
	"testing"
)

func TestReadStringValidation(t *testing.T) {
	tests := []struct {
		name       string
		buf        []byte
		wantStr    string
		wantConsum int
	}{
		{
			name:       "empty buffer",
			buf:        []byte{},
			wantStr:    "",
			wantConsum: 0,
		},
		{
			name:       "buffer too small for length prefix",
			buf:        []byte{0x01, 0x02},
			wantStr:    "",
			wantConsum: 0,
		},
		{
			name:       "zero length string",
			buf:        []byte{0x00, 0x00, 0x00, 0x00},
			wantStr:    "",
			wantConsum: 4,
		},
		{
			name:       "negative length",
			buf:        []byte{0xFF, 0xFF, 0xFF, 0xFF}, // -1 in little-endian int32
			wantStr:    "",
			wantConsum: 4,
		},
		{
			name:       "length exceeds buffer",
			buf:        []byte{0x0A, 0x00, 0x00, 0x00, 'h', 'i'}, // length=10 but only 2 data bytes
			wantStr:    "",
			wantConsum: 4,
		},
		{
			name:       "valid string",
			buf:        []byte{0x05, 0x00, 0x00, 0x00, 'h', 'e', 'l', 'l', 'o'},
			wantStr:    "hello",
			wantConsum: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotConsum := readString(tt.buf)
			if gotStr != tt.wantStr {
				t.Errorf("readString() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotConsum != tt.wantConsum {
				t.Errorf("readString() consumed = %d, want %d", gotConsum, tt.wantConsum)
			}
		})
	}
}

func TestWriteReadStringRoundTrip(t *testing.T) {
	tests := []string{"", "hello", "test string", "日本語"}

	for _, s := range tests {
		t.Run(s, func(t *testing.T) {
			buf := make([]byte, 4+len(s))
			n := writeString(buf, s)
			if n != 4+len(s) {
				t.Errorf("writeString() = %d, want %d", n, 4+len(s))
			}

			gotStr, gotConsum := readString(buf)
			if gotStr != s {
				t.Errorf("round-trip: got %q, want %q", gotStr, s)
			}
			if gotConsum != n {
				t.Errorf("round-trip consumed: got %d, want %d", gotConsum, n)
			}
		})
	}
}

func TestWriteReadFloat64RoundTrip(t *testing.T) {
	tests := []float64{0, 1.0, -1.0, 3.14159, 1e100, -1e-100}

	for _, v := range tests {
		buf := make([]byte, 8)
		writeFloat64(buf, v)
		got := readFloat64(buf)
		if got != v {
			t.Errorf("round-trip float64: got %v, want %v", got, v)
		}
	}
}

func TestWriteReadBoolRoundTrip(t *testing.T) {
	for _, v := range []bool{true, false} {
		buf := make([]byte, 1)
		writeBool(buf, v)
		got := readBool(buf)
		if got != v {
			t.Errorf("round-trip bool: got %v, want %v", got, v)
		}
	}
}
