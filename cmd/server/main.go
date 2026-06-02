package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/yourname/code-snippet-explainer/internal/model"
	"github.com/yourname/code-snippet-explainer/internal/service"
)

var pageTemplate = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Code Snippet Explainer</title>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <style>
      body { font-family: system-ui, sans-serif; margin: 2rem; color: #111; }
      main { max-width: 900px; margin: 0 auto; }
      label { display: block; margin-bottom: 0.75rem; font-weight: 600; }
      textarea, select, input { width: 100%; font: 1rem ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; }
      textarea { min-height: 220px; padding: 0.75rem; border: 1px solid #cbd5e1; border-radius: 0.5rem; }
      select, input { padding: 0.65rem 0.85rem; border: 1px solid #cbd5e1; border-radius: 0.5rem; background: #fff; }
      button { padding: 0.75rem 1rem; border: 1px solid #cbd5e1; border-radius: 0.5rem; background: #1f2937; color: #fff; cursor: pointer; font-weight: 700; }
      #explanation { margin-top: 1.5rem; min-height: 12rem; border: 1px solid #e2e8f0; padding: 1rem; border-radius: 0.75rem; background: #f8fafc; white-space: pre-wrap; }
      .htmx-indicator { display: none; }
      .htmx-request .htmx-indicator { display: inline; }
    </style>
  </head>
  <body>
    <main>
      <h1>Code Snippet Explainer</h1>
      <form id="explain-form" method="post" hx-post="/explain" hx-target="#explanation" hx-swap="innerHTML" hx-include="[name='code'],[name='language'],[name='mode']" hx-indicator="#spinner">
        <label>
          Language override:<br />
          <select name="language">
            <option value="">Auto-detect language</option>
            <option value="JavaScript">JavaScript</option>
            <option value="TypeScript">TypeScript</option>
            <option value="Python">Python</option>
            <option value="Go">Go</option>
            <option value="Rust">Rust</option>
            <option value="Java">Java</option>
            <option value="C">C</option>
            <option value="C++">C++</option>
            <option value="SQL">SQL</option>
            <option value="Bash">Bash</option>
          </select>
        </label>

        <label>
          Mode:<br />
          <select name="mode">
            <option value="summary">Summary</option>
            <option value="line-by-line">Line by line</option>
          </select>
        </label>

        <label>
          Code:<br />
          <textarea id="code-input" name="code" rows="12" placeholder="Paste your code here..." required></textarea>
        </label>

        <button type="submit">Explain</button>
        <span id="spinner" class="htmx-indicator">Loading…</span>
      </form>

      <section id="explanation">
        <em>Explanation output will appear here.</em>
      </section>
    </main>
  </body>
</html>`))

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTemplate.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func explainHandler(explainer *service.Explainer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		code := strings.TrimSpace(r.FormValue("code"))
		language := strings.TrimSpace(r.FormValue("language"))
		mode := strings.TrimSpace(r.FormValue("mode"))

		if code == "" || mode == "" {
			http.Error(w, "Missing form fields", http.StatusBadRequest)
			return
		}

		if language == "" {
			language = service.DetectLanguage(code)
			if language == "unknown" {
				language = ""
			}
		}

		req := model.ExplainRequest{Code: code, Language: language, Mode: mode}
		if err := service.ValidateRequest(req); err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "<p class=\"error\">%s</p>", html.EscapeString(err.Error()))
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		fmt.Fprint(w, `<div class="explanation">`)
		flusher.Flush()

		err := explainer.ExplainStream(r.Context(), code, language, mode, func(chunk string) error {
			fmt.Fprintf(w, "<p>%s</p>", html.EscapeString(chunk))
			flusher.Flush()
			return nil
		})
		if err != nil {
			fmt.Fprintf(w, "<p class=\"error\">%s</p>", html.EscapeString(err.Error()))
			flusher.Flush()
		}

		fmt.Fprint(w, `</div>`)
		flusher.Flush()
	}
}

func main() {
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	explainer := service.NewExplainer(ollamaURL)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/explain", explainHandler(explainer))

	addr := ":8081"
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
