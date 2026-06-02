// Package service contains business logic for the code snippet explainer.
//
// DetectLanguage attempts to identify the programming language of a code snippet
// using heuristics (keyword matching, file extension hints, syntax patterns).
// It returns one of the ten supported languages:
//   JavaScript, TypeScript, Python, Go, Rust, Java, C, C++, SQL, Bash
// If detection is not confident, it returns "unknown".
//
// Rules:
//   - Check for shebang lines first (e.g. #!/usr/bin/env python)
//   - Match distinctive keywords: "func " and ":=" suggest Go; "fn " and "let mut" suggest Rust
//   - SQL detection: presence of SELECT/INSERT/UPDATE/CREATE keywords (case-insensitive)
//   - Bash: shebang or heavy use of $VAR syntax and commands like echo, grep, awk
//   - TypeScript vs JavaScript: presence of type annotations or "interface " keyword
//
// DetectLanguage(code string) string

package service

import "strings"

func DetectLanguage(code string) string {
	code = strings.TrimSpace(code)
	lines := strings.Split(code, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		if strings.HasPrefix(firstLine, "#!") {
			if strings.Contains(firstLine, "python") {
				return "Python"
			} else if strings.Contains(firstLine, "bash") || strings.Contains(firstLine, "sh") {
				return "Bash"
			}
		}
	}

	if strings.Contains(code, "func ") && strings.Contains(code, ":=") {
		return "Go"
	}
	if strings.Contains(code, "fn ") && strings.Contains(code, "let mut") {
		return "Rust"
	}
	if strings.Contains(strings.ToUpper(code), "SELECT ") ||
		strings.Contains(strings.ToUpper(code), "INSERT ") ||
		strings.Contains(strings.ToUpper(code), "UPDATE ") ||
		strings.Contains(strings.ToUpper(code), "CREATE ") {
		return "SQL"
	}
	if strings.Contains(code, "interface ") || strings.Contains(code, ": ") {
		return "TypeScript"
	}
	if strings.Contains(code, "echo ") || strings.Contains(code, "grep ") || strings.Contains(code, "awk ") {
		return "Bash"
	}
	return "unknown"
}
