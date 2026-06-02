# Code Snippet Explainer

A lightweight Go web application that takes any code snippet and returns a plain-English explanation using a local Llama 3.2 model via Ollama. No cloud API keys required.

---

## What it does

Paste any code snippet into the text area, choose an explanation mode, and click **Explain**. The explanation streams into the page in real time without a page refresh.

Two explanation modes are supported:

- **Summary** — a 2–4 sentence description of what the code does overall
- **Line by line** — a numbered breakdown of each logical block

Ten programming languages are supported at launch: JavaScript, TypeScript, Python, Go, Rust, Java, C, C++, SQL, and Bash. Language is auto-detected from the snippet but can be overridden manually.

---

## Tech stack

| Layer | Technology |
|---|---|
| Language | Go |
| HTTP server | `net/http` (standard library) |
| Templating | `html/template` (standard library) |
| Frontend interaction | htmx via CDN |
| AI backend | Ollama running Llama 3.2 locally |
| Streaming | `http.Flusher` + Ollama streaming API |

---

## Project structure
