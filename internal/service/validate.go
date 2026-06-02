// ValidateRequest checks an ExplainRequest before it is passed to the explainer.
//
// ValidateRequest(req model.ExplainRequest) error
//
// Rules (return a descriptive error string for each):
//   - Code must not be empty → "code is required"
//   - Code must not exceed 300 lines → "code exceeds 300 line limit"
//   - Mode must be "summary" or "line-by-line" → "mode must be summary or line-by-line"
//   - If Language is non-empty, it must be one of model.SupportedLanguages → "unsupported language"
//
// Count lines by splitting on "\n". Trim trailing whitespace before counting.

package service

import (
	"errors"
	"strings"

	"github.com/yourname/code-snippet-explainer/internal/model"
)

func ValidateRequest(req model.ExplainRequest) error {
	if strings.TrimSpace(req.Code) == "" {
		return errors.New("code is required")
	}
	lines := strings.Split(strings.TrimSpace(req.Code), "\n")
	if len(lines) > 300 {
		return errors.New("code exceeds 300 line limit")
	}
	if req.Mode != "summary" && req.Mode != "line-by-line" {
		return errors.New("mode must be summary or line-by-line")
	}
	if req.Language != "" {
		supported := false
		for _, lang := range model.SupportedLanguages {
			if req.Language == lang {
				supported = true
				break
			}
		}
		if !supported {
			return errors.New("unsupported language")
		}
	}
	return nil
}
