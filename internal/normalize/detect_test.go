package normalize_test

import (
	"feedforge/internal/normalize"
	"testing"
)

func TestDetectType(t *testing.T) {
	tests := []struct {
		Name    string
		Input   string
		Want    normalize.Type
		WantErr bool
	}{
		{Name: "http url", Input: "http://evil.com/payload.exe", Want: normalize.TypeURL},
		{Name: "https url", Input: "https://evil.com", Want: normalize.TypeURL},
		{Name: "ipv4", Input: "192.168.1.1", Want: normalize.TypeIP},
		{Name: "ipv6", Input: "2001:db8::1", Want: normalize.TypeIPv6},
		{Name: "md5", Input: "d41d8cd98f00b204e9800998ecf8427e", Want: normalize.TypeHash},
		{Name: "domain", Input: "evil.example.com", Want: normalize.TypeDomain},
		{Name: "email", Input: "user@evil.com", Want: normalize.TypeEmail},
		{Name: "empty", Input: "", WantErr: true},
		{Name: "garbage", Input: "not-anything-valid?", WantErr: true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := normalize.DetectType(test.Input)

			if test.WantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != test.Want {
				t.Errorf("got %q, want %q", got, test.Want)
			}
		})
	}
}
