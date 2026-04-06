// Package layers implements the Ralph ↔ Layers (TypeScript) contract (see docs/LAYERS_SPEC.md).
package layers

// RetrieveRequest mirrors layers/src/contracts/v1.ts RetrieveRequest.
type RetrieveRequest struct {
	ProjectRoot string `json:"projectRoot"`
	DataDir     string `json:"dataDir,omitempty"`
	Query       struct {
		Text      string `json:"text"`
		Category  string `json:"category,omitempty"`
		FeatureID int    `json:"featureId,omitempty"`
	} `json:"query"`
	Options *RetrieveOptions `json:"options,omitempty"`
}

// RetrieveOptions mirrors optional retrieve options.
type RetrieveOptions struct {
	TopK                int     `json:"topK,omitempty"`
	MaxTokens           int     `json:"maxTokens,omitempty"`
	MmrLambda           float64 `json:"mmrLambda,omitempty"`
	EmbeddingFallbackOk *bool   `json:"embeddingFallbackOk,omitempty"`
}

// RetrieveResponse mirrors a successful retrieve stdout JSON.
type RetrieveResponse struct {
	OK           bool   `json:"ok"`
	ContextBlock string `json:"contextBlock"`
	Memories     []any  `json:"memories,omitempty"`
	Meta         any    `json:"meta,omitempty"`
}

// ErrorBody is returned when ok=false.
type ErrorBody struct {
	OK    bool `json:"ok"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// RecordRequest mirrors RecordRequest in the TS contract.
type RecordRequest struct {
	ProjectRoot string        `json:"projectRoot"`
	DataDir     string        `json:"dataDir,omitempty"`
	Entries     []RecordEntry `json:"entries"`
}

// RecordEntry is one memory row for record.
type RecordEntry struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Category  string `json:"category,omitempty"`
	FeatureID *int   `json:"featureId,omitempty"`
	Source    string `json:"source"`
}

// AppendRunRequest mirrors AppendRunRequest.
type AppendRunRequest struct {
	ProjectRoot string        `json:"projectRoot"`
	DataDir     string        `json:"dataDir,omitempty"`
	RunLog      string        `json:"runLog,omitempty"`
	Event       RunEventInput `json:"event"`
}

// RunEventInput mirrors RunEventInput / RunEvent in TS.
type RunEventInput struct {
	SessionKey string         `json:"sessionKey"`
	Kind       string         `json:"kind"`
	Iteration  int            `json:"iteration,omitempty"`
	FeatureID  int            `json:"featureId,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
	TS         string         `json:"ts,omitempty"`
}

// CompactRequest mirrors CompactRequest.
type CompactRequest struct {
	ProjectRoot string `json:"projectRoot"`
	DataDir     string `json:"dataDir,omitempty"`
	RunLog      string `json:"runLog,omitempty"`
	Snapshot    string `json:"snapshot,omitempty"`
	MaxEvents   int    `json:"maxEvents,omitempty"`
	MaxBytes    int    `json:"maxBytes,omitempty"`
}
