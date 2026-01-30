package handler

import (
	"fmt"
	"strings"
	"testing"

	"familyshare/internal/pipeline"
)

func TestFriendlyUploadErrorMapping(t *testing.T) {
	maxPerFile := int64(25 << 20)

	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "not image",
			err:  pipeline.ErrNotAnImage,
			want: "Unsupported file type",
		},
		{
			name: "decode failed",
			err:  fmt.Errorf("wrap: %w", pipeline.ErrDecodeFailed),
			want: "couldn't read",
		},
		{
			name: "invalid dimensions",
			err:  pipeline.ErrInvalidDimensions,
			want: "8000",
		},
		{
			name: "too large",
			err:  errUploadTooLarge,
			want: "25MB",
		},
		{
			name: "default",
			err:  fmt.Errorf("boom"),
			want: "Upload failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := friendlyUploadError(tc.err, maxPerFile)
			if !strings.Contains(strings.ToLower(msg), strings.ToLower(tc.want)) {
				t.Fatalf("expected message containing %q, got %q", tc.want, msg)
			}
		})
	}
}
