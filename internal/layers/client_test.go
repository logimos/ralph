package layers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Retrieve_HTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/retrieve" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&map[string]any{})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":           true,
			"contextBlock": "[MEMORY CONTEXT]\n- ok\n[END MEMORY CONTEXT]\n",
			"memories":     []any{},
			"meta":         map[string]any{"ftsOnly": true, "truncated": false},
		})
	}))
	defer ts.Close()

	c := &Client{BaseURL: ts.URL}
	ctx := context.Background()
	out, err := c.Retrieve(ctx, RetrieveRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if out.ContextBlock == "" {
		t.Fatal("expected contextBlock")
	}
}
