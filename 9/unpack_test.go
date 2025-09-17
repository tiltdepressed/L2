package main

import "testing"

func TestStringUnpack(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "simple case",
			input:   "a4bc2d5e",
			want:    "aaaabccddddde",
			wantErr: false,
		},
		{
			name:    "no digits",
			input:   "abcd",
			want:    "abcd",
			wantErr: false,
		},
		{
			name:    "invalid string (only digits)",
			input:   "45",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "escaped digits",
			input:   `qwe\4\5`,
			want:    "qwe45",
			wantErr: false,
		},
		{
			name:    "escaped digit followed by repetition",
			input:   `qwe\45`,
			want:    "qwe44444",
			wantErr: false,
		},
		{
			name:    "escaped backslash",
			input:   `qwe\\5`,
			want:    `qwe\\\\\`,
			wantErr: false,
		},
		{
			name:    "single character",
			input:   "a",
			want:    "a",
			wantErr: false,
		},
		{
			name:    "repetition with zero",
			input:   "a0b1c0d",
			want:    "bd",
			wantErr: false,
		},
		{
			name:    "multibyte characters (Unicode)",
			input:   "🙂2🙁3",
			want:    "🙂🙂🙁🙁🙁",
			wantErr: false,
		},
		{
			name:    "invalid string (digit at start)",
			input:   "1a",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid string (trailing backslash)",
			input:   `abc\`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid string (multiple digits after letter)",
			input:   "a12",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := stringUnpack(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("stringUnpack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("stringUnpack() = %q, want %q", got, tt.want)
			}
		})
	}
}
