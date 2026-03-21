package web

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubComponent struct {
	content string
	err     error
}

func (s stubComponent) Render(_ context.Context, w io.Writer) error {
	if s.err != nil {
		return s.err
	}
	_, err := w.Write([]byte(s.content))
	return err
}

func TestRenderTempl_Success(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	RenderTempl(w, r, logger, stubComponent{content: "<h1>OK</h1>"})

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type text/html; charset=utf-8, got %q", ct)
	}
	if w.Body.String() != "<h1>OK</h1>" {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestRenderTempl_Error(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	RenderTempl(w, r, logger, stubComponent{err: fmt.Errorf("render failed")})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
