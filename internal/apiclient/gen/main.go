// Command gen downloads the Porkbun OpenAPI spec, repairs known upstream defects,
// and runs oapi-codegen on the result.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const specURL = "https://porkbun.com/api/json/v3/spec"

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Printf("fetching spec from %s", specURL)
	resp, err := http.Get(specURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching spec: %s", resp.Status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("downloaded %d bytes", len(raw))

	var spec any
	if err := json.Unmarshal(raw, &spec); err != nil {
		return fmt.Errorf("parsing spec: %w", err)
	}
	log.Print("repairing spec (empty `properties` and `items` arrays)")
	fixEmptyArrays(spec)
	log.Print("repairing spec (duplicate parameter refs)")
	dedupeParameters(spec)

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(spec); err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "porkbun-spec-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	specPath := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(specPath, buf.Bytes(), 0o600); err != nil {
		return err
	}

	log.Printf("running oapi-codegen (spec: %s)", specPath)
	cmd := exec.Command("go", "tool", "oapi-codegen", "-config", "config.yaml", specPath)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("oapi-codegen: %w", err)
	}
	log.Print("generated client successfully")
	return nil
}

// fixEmptyArrays replaces `"properties": []` and `"items": []` (empty PHP arrays
// serialized as JSON arrays) with an empty object, as required by OpenAPI.
func fixEmptyArrays(v any) {
	switch v := v.(type) {
	case map[string]any:
		for k, child := range v {
			if arr, ok := child.([]any); ok && (k == "properties" || k == "items") && len(arr) == 0 {
				log.Printf("  fixed empty array at %q", k)
				v[k] = map[string]any{}
				continue
			}
			fixEmptyArrays(child)
		}
	case []any:
		for _, child := range v {
			fixEmptyArrays(child)
		}
	}
}

// dedupeParameters drops repeated `$ref` entries from each operation's
// `parameters`, which upstream sometimes lists twice (e.g. IdempotencyKey).
func dedupeParameters(v any) {
	switch v := v.(type) {
	case map[string]any:
		if params, ok := v["parameters"].([]any); ok {
			seen := map[string]bool{}
			kept := params[:0]
			for _, p := range params {
				if ref, ok := p.(map[string]any)["$ref"].(string); ok {
					if seen[ref] {
						log.Printf("  dropped duplicate parameter %q", ref)
						continue
					}
					seen[ref] = true
				}
				kept = append(kept, p)
			}
			v["parameters"] = kept
		}
		for _, child := range v {
			dedupeParameters(child)
		}
	case []any:
		for _, child := range v {
			dedupeParameters(child)
		}
	}
}
