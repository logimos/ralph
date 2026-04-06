package layers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const defaultCLI = "layers"

// Client calls the Layers TS service via CLI subprocess or HTTP.
type Client struct {
	// Command is the `layers` binary (e.g. "layers" or "node"). Empty uses defaultCLI + PATH.
	Command string
	// BaseURL is e.g. "http://127.0.0.1:7847". If set, HTTP is used instead of CLI.
	BaseURL string
	// HTTPClient used when BaseURL is set.
	HTTPClient *http.Client
	// Env is passed to the subprocess (e.g. OPENAI_API_KEY); may be nil.
	Env []string
}

func (c *Client) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 120 * time.Second}
}

func (c *Client) useHTTP() bool {
	return strings.TrimSpace(c.BaseURL) != ""
}

func (c *Client) cliName() string {
	if strings.TrimSpace(c.Command) != "" {
		return strings.TrimSpace(c.Command)
	}
	return defaultCLI
}

// postJSON runs a v1 CLI subcommand or HTTP POST to /v1/<subcommand>.
func (c *Client) postJSON(ctx context.Context, subcommand string, body any, out any) error {
	if c.useHTTP() {
		path := "/v1/" + strings.TrimPrefix(strings.TrimSpace(subcommand), "/")
		return c.postHTTP(ctx, path, body, out)
	}
	return c.postCLI(ctx, subcommand, body, out)
}

func (c *Client) postHTTP(ctx context.Context, path string, body any, out any) error {
	base := strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var eb ErrorBody
		if json.Unmarshal(raw, &eb) == nil && eb.Error.Message != "" {
			return fmt.Errorf("layers http %s: %s", resp.Status, eb.Error.Message)
		}
		return fmt.Errorf("layers http %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("layers http decode: %w", err)
	}
	return nil
}

func (c *Client) postCLI(ctx context.Context, subcommand string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	name, extra := parseCommandLine(c.cliName())
	if name == "" {
		name = defaultCLI
	}
	args := append(append([]string{}, extra...), "v1", subcommand)
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(payload)
	if len(c.Env) > 0 {
		cmd.Env = c.Env
	} else {
		cmd.Env = os.Environ()
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("layers cli %s: %w (%s)", subcommand, err, msg)
	}
	raw := stdout.Bytes()
	var eb ErrorBody
	if json.Unmarshal(raw, &eb) == nil && !eb.OK && eb.Error.Message != "" {
		return fmt.Errorf("layers: %s", eb.Error.Message)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("layers cli decode: %w", err)
	}
	return nil
}

// Retrieve runs v1 retrieve.
func (c *Client) Retrieve(ctx context.Context, req RetrieveRequest) (*RetrieveResponse, error) {
	var out RetrieveResponse
	if err := c.postJSON(ctx, "retrieve", req, &out); err != nil {
		return nil, err
	}
	if !out.OK {
		return nil, errors.New("layers retrieve: ok=false")
	}
	return &out, nil
}

// Record runs v1 record.
func (c *Client) Record(ctx context.Context, req RecordRequest) error {
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.postJSON(ctx, "record", req, &out); err != nil {
		return err
	}
	if !out.OK {
		return errors.New("layers record: ok=false")
	}
	return nil
}

// AppendRun runs v1 append-run.
func (c *Client) AppendRun(ctx context.Context, req AppendRunRequest) error {
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.postJSON(ctx, "append-run", req, &out); err != nil {
		return err
	}
	if !out.OK {
		return errors.New("layers append-run: ok=false")
	}
	return nil
}

// Compact runs v1 compact.
func (c *Client) Compact(ctx context.Context, req CompactRequest) error {
	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.postJSON(ctx, "compact", req, &out); err != nil {
		return err
	}
	if !out.OK {
		return errors.New("layers compact: ok=false")
	}
	return nil
}
