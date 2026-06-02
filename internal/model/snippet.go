// Package model defines the request and response types for the code snippet explainer.
//
// ExplainRequest is the form values sent by the client:
//   - Code     string — the raw code snippet (required, max 300 lines)
//   - Language string — programming language override (optional); if empty, auto-detect
//   - Mode     string — explanation mode: "summary" or "line-by-line" (required)
//
// ExplainResponse is returned to the handler after the AI call:
//   - Explanation string — the plain-English explanation
//   - Language    string — the detected or overridden language
//   - Mode        string — the mode that was used
//   - Error       string — non-empty if something went wrong
//
// SupportedLanguages lists the ten languages supported at launch:
// JavaScript, TypeScript, Python, Go, Rust, Java, C, C++, SQL, Bash
// Define it as a package-level var so handlers can validate against it.
package model

var SupportedLanguages = []string{
	"JavaScript",
	"TypeScript",
	"Python",
	"Go",
	"Rust",
	"Java",
	"C",
	"C++",
	"SQL",
	"Bash",
}

type ExplainRequest struct {
	Code     string `json:"code"`
	Language string `json:"language,omitempty"`
	Mode     string `json:"mode"`
}
type ExplainResponse struct {
	Explanation string `json:"explanation"`
	Language    string `json:"language"`
	Mode        string `json:"mode"`
	Error       string `json:"error,omitempty"`
}
