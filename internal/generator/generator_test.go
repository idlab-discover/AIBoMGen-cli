package generator

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/builder"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
	"github.com/idlab-discover/AIBoMGen-cli/internal/ui"
)

type failingBuilder struct {
	err error
}

func (f *failingBuilder) Build(builder.BuildContext) (*cdx.BOM, error) {
	return nil, f.err
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestBuildPerDiscovery_FetchesMetadataAndBuilds(t *testing.T) {
	responses := []string{
		`{"id":"hf-alpha","modelId":"hf-alpha","author":"org","pipeline_tag":"tag","library_name":"lib","tags":["t1"],"license":"mit","sha":"abc","downloads":1,"likes":1,"lastModified":"2024-01-01","createdAt":"2023-01-01","private":false,"usedStorage":1,"cardData":{"license":"mit"}}`,
		`{"id":"hf-beta","modelId":"hf-beta","author":"org","pipeline_tag":"tag","library_name":"lib","tags":["t2"],"license":"apache","sha":"def","downloads":2,"likes":2,"lastModified":"2024-02-01","createdAt":"2023-02-01","private":false,"usedStorage":2,"cardData":{"license":"apache"}}`,
	}

	origTransport := http.DefaultTransport
	var paths []string
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		idx := len(paths)
		if idx >= len(responses) {
			t.Fatalf("unexpected request #%d to %s", idx+1, req.URL)
		}
		if got, want := req.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Fatalf("authorization header = %q, want %q", got, want)
		}
		paths = append(paths, req.URL.Path)
		body := io.NopCloser(strings.NewReader(responses[idx]))
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	discoveries := []scanner.Discovery{
		{ID: "org-model", Path: "model.py"},
		{Name: "beta"},
		{},
	}

	got, err := BuildPerDiscovery(discoveries, "test-token", 0)
	if err != nil {
		t.Fatalf("BuildPerDiscovery() error = %v", err)
	}
	if len(got) != len(discoveries) {
		t.Fatalf("results len = %d, want %d", len(got), len(discoveries))
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 fetches, got %d", len(paths))
	}
	if got[0].BOM == nil || got[0].BOM.Metadata == nil || got[0].BOM.Metadata.Component == nil {
		t.Fatalf("first BOM missing metadata/component")
	}
	if got[0].BOM.Metadata.Component.Name != "hf-alpha" {
		t.Fatalf("first component name = %q, want hf-alpha", got[0].BOM.Metadata.Component.Name)
	}
	if got[1].Discovery.Name != "beta" {
		t.Fatalf("second discovery preserved name, got %q", got[1].Discovery.Name)
	}
	if got[2].BOM.Metadata.Component.Name != "model" {
		t.Fatalf("third component default name = %q, want model", got[2].BOM.Metadata.Component.Name)
	}
	if !strings.Contains(paths[1], "beta") {
		t.Fatalf("second request path %q missing beta", paths[1])
	}
}

func TestBuildPerDiscovery_FetchErrorStillBuilds(t *testing.T) {
	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("boom")
	})
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	discoveries := []scanner.Discovery{{ID: "err-model"}}
	got, err := BuildPerDiscovery(discoveries, "", 5*time.Second)
	if err != nil {
		t.Fatalf("BuildPerDiscovery() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].BOM == nil || got[0].BOM.Metadata == nil || got[0].BOM.Metadata.Component == nil {
		t.Fatalf("result bom missing metadata/component")
	}
	if got[0].BOM.Metadata.Component.Name != "err-model" {
		t.Fatalf("component name = %q, want err-model", got[0].BOM.Metadata.Component.Name)
	}
}

func TestLogfWritesWithConfiguredLogger(t *testing.T) {
	ui.Init(true)

	var buf bytes.Buffer
	SetLogger(&buf)
	t.Cleanup(func() { SetLogger(nil) })

	logf("model-x", "hello %s", "world")

	got := buf.String()
	for _, want := range []string{"Generator:", "model=model-x", "hello world"} {
		if !strings.Contains(got, want) {
			t.Fatalf("log output %q missing %q", got, want)
		}
	}
}

func TestBuildPerDiscovery_PropagatesBuilderError(t *testing.T) {
	origFactory := newBOMBuilder
	newBOMBuilder = func() bomBuilder {
		return &failingBuilder{err: errors.New("builder boom")}
	}
	t.Cleanup(func() { newBOMBuilder = origFactory })

	_, err := BuildPerDiscovery([]scanner.Discovery{{}}, "", time.Second)
	if err == nil || !strings.Contains(err.Error(), "builder boom") {
		t.Fatalf("expected builder error, got %v", err)
	}
}
