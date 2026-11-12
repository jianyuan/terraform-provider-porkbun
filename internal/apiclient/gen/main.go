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
	log.Print("repairing spec (empty `properties` arrays)")
	fixEmptyProperties(spec)

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

// fixEmptyProperties replaces `"properties": []` (an empty PHP array serialized
// as a JSON array) with an empty object, as required by OpenAPI.
func fixEmptyProperties(v any) {
	switch v := v.(type) {
	case map[string]any:
		for k, child := range v {
			if arr, ok := child.([]any); ok && k == "properties" && len(arr) == 0 {
				log.Printf("  fixed empty array at %q", k)
				v[k] = map[string]any{}
				continue
			}
			fixEmptyProperties(child)
		}
	case []any:
		for _, child := range v {
			fixEmptyProperties(child)
		}
	}
}
