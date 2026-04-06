package layers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_postHTTP_errorBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": false,
			"error": map[string]string{
				"code":    "INVALID_REQUEST",
				"message": "bad thing",
			},
		})
	}))
	defer ts.Close()

	c := &Client{BaseURL: ts.URL}
	ctx := context.Background()
	var out struct{}
	err := c.postHTTP(ctx, "/v1/record", map[string]string{"x": "y"}, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	es := err.Error()
	if !strings.Contains(es, "400") || !strings.Contains(es, "bad thing") {
		t.Fatalf("unexpected: %v", err)
	}
}
