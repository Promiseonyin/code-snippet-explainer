package service

import (
	"strings"
	"testing"

	"github.com/yourname/code-snippet-explainer/internal/model"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     model.ExplainRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty code returns error",
			req:     model.ExplainRequest{Code: "", Mode: "summary"},
			wantErr: true,
			errMsg:  "code is required",
		},
		{
			name: "code over 300 lines returns error",
			req: model.ExplainRequest{
				Code: strings.Repeat("x\n", 301),
				Mode: "summary",
			},
			wantErr: true,
			errMsg:  "code exceeds 300 line limit",
		},
		{
			name:    "invalid mode returns error",
			req:     model.ExplainRequest{Code: "println(\"hi\")", Mode: "detailed"},
			wantErr: true,
			errMsg:  "mode must be summary or line-by-line",
		},
		{
			name:    "unsupported language returns error",
			req:     model.ExplainRequest{Code: "print(\"hi\")", Mode: "summary", Language: "Pascal"},
			wantErr: true,
			errMsg:  "unsupported language",
		},
		{
			name:    "valid request returns nil",
			req:     model.ExplainRequest{Code: "println(\"hi\")", Mode: "summary", Language: "Go"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Fatalf("ValidateRequest() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}
