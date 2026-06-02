package service

import "testing"

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "Go snippet returns Go",
			code: "package main\nfunc main() {\n    x := 10\n}\n",
			want: "Go",
		},
		{
			name: "Python shebang returns Python",
			code: "#!/usr/bin/env python\nprint(\"hello\")\n",
			want: "Python",
		},
		{
			name: "SQL with SELECT returns SQL",
			code: "SELECT id, name FROM users WHERE active = 1;\n",
			want: "SQL",
		},
		{
			name: "Bash shebang returns Bash",
			code: "#!/bin/bash\necho hello\n",
			want: "Bash",
		},
		{
			name: "unrecognisable snippet returns unknown",
			code: "some random text with no code patterns\n",
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectLanguage(tt.code)
			if got != tt.want {
				t.Fatalf("DetectLanguage() = %q, want %q", got, tt.want)
			}
		})
	}
}
