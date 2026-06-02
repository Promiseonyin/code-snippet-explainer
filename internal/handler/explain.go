// ExplainHandler handles POST /explain requests.
//
// NewExplainHandler(svc *service.Explainer, detect func(string) string) http.HandlerFunc
//
// Request flow:
//   1. Reject non-POST methods with 405.
//   2. Decode JSON body into model.ExplainRequest. Return 400 on decode error.
//   3. If Language is empty, call detect(req.Code) to fill it in.
//   4. Call service.ValidateRequest(req). Return 422 with error message on failure.
//   5. Call svc.Explain(r.Context(), req.Code, req.Language, req.Mode).
//      Return 502 with a safe error message if the AI call fails.
//   6. Write a 200 JSON response: model.ExplainResponse{
//        Explanation: result,
//        Language:    req.Language,
//        Mode:        req.Mode,
//      }
//
// All error responses are JSON: {"error": "message here"}
// Set Content-Type: application/json on every response.

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yourname/code-snippet-explainer/internal/model"
	"github.com/yourname/code-snippet-explainer/internal/service"
)

func NewExplainHandler(svc *service.Explainer, detect func(string) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		var req model.ExplainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		if req.Language == "" {
			req.Language = detect(req.Code)
		}
		if err := service.ValidateRequest(req); err != nil {
			http.Error(w, `{"error": "`+err.Error()+`"}`, http.StatusUnprocessableEntity)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, `{"error": "streaming unsupported"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		fmt.Fprint(w, `<div class="explanation">`)
		flusher.Flush()

		err := svc.ExplainStream(r.Context(), req.Code, req.Language, req.Mode, func(chunk string) error {
			fmt.Fprint(w, "<span>"+chunk+"</span>")
			flusher.Flush()
			return nil
		})
		if err != nil {
			fmt.Fprint(w, `<span class="error">stream error</span>`)
			flusher.Flush()
			return
		}

		fmt.Fprint(w, `</div>`)
		flusher.Flush()
	}
}
