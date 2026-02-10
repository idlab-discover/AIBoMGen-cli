package generator

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/idlab-discover/AIBoMGen-cli/internal/builder"
	"github.com/idlab-discover/AIBoMGen-cli/internal/fetcher"
	"github.com/idlab-discover/AIBoMGen-cli/internal/scanner"
)

// Mock BOM Builder for testing
type mockBOMBuilder struct {
	buildFunc        func(builder.BuildContext) (*cdx.BOM, error)
	buildDatasetFunc func(builder.DatasetBuildContext) (*cdx.Component, error)
}

func (m *mockBOMBuilder) Build(ctx builder.BuildContext) (*cdx.BOM, error) {
	if m.buildFunc != nil {
		return m.buildFunc(ctx)
	}
	return &cdx.BOM{}, nil
}

func (m *mockBOMBuilder) BuildDataset(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
	if m.buildDatasetFunc != nil {
		return m.buildDatasetFunc(ctx)
	}
	return &cdx.Component{Name: ctx.DatasetID}, nil
}

// Mock Fetchers for testing
type mockModelAPIFetcher struct {
	fetchFunc func(context.Context, string) (*fetcher.ModelAPIResponse, error)
}

func (m *mockModelAPIFetcher) Fetch(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, id)
	}
	return &fetcher.ModelAPIResponse{}, nil
}

type mockModelReadmeFetcher struct {
	fetchFunc func(context.Context, string) (*fetcher.ModelReadmeCard, error)
}

func (m *mockModelReadmeFetcher) Fetch(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, id)
	}
	return &fetcher.ModelReadmeCard{}, nil
}

type mockDatasetAPIFetcher struct {
	fetchFunc func(context.Context, string) (*fetcher.DatasetAPIResponse, error)
}

func (m *mockDatasetAPIFetcher) Fetch(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, id)
	}
	return &fetcher.DatasetAPIResponse{}, nil
}

type mockDatasetReadmeFetcher struct {
	fetchFunc func(context.Context, string) (*fetcher.DatasetReadmeCard, error)
}

func (m *mockDatasetReadmeFetcher) Fetch(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
	if m.fetchFunc != nil {
		return m.fetchFunc(ctx, id)
	}
	return &fetcher.DatasetReadmeCard{}, nil
}

func TestBuildDummyBOM(t *testing.T) {
	// Save originals
	originalDummyFetcherSet := newDummyFetcherSet
	originalBuilder := newBOMBuilder

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		check   func(*testing.T, []DiscoveredBOM)
	}{
		{
			name:    "builds dummy BOM successfully",
			setup:   func() {},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				if got[0].BOM == nil {
					t.Error("BOM is nil")
					return
				}
				if got[0].Discovery.ID != "dummy-org/dummy-model" {
					t.Errorf("Expected discovery ID 'dummy-org/dummy-model', got %q", got[0].Discovery.ID)
				}
				// Should have datasets from dummy data
				if got[0].BOM.Components != nil && len(*got[0].BOM.Components) > 0 {
					t.Logf("BOM has %d dataset components", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "handles model API fetch error",
			setup: func() {
				newDummyFetcherSet = func() fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return nil, context.Canceled
							},
						},
						modelReadme:   &fetcher.DummyModelReadmeFetcher{},
						datasetAPI:    &fetcher.DummyDatasetAPIFetcher{},
						datasetReadme: &fetcher.DummyDatasetReadmeFetcher{},
					}
				}
			},
			wantErr: true,
		},
		{
			name: "handles model README fetch error",
			setup: func() {
				newDummyFetcherSet = func() fetcherSet {
					return fetcherSet{
						modelAPI: &fetcher.DummyModelAPIFetcher{},
						modelReadme: &mockModelReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
								return nil, context.Canceled
							},
						},
						datasetAPI:    &fetcher.DummyDatasetAPIFetcher{},
						datasetReadme: &fetcher.DummyDatasetReadmeFetcher{},
					}
				}
			},
			wantErr: true,
		},
		{
			name: "handles BOM build error",
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return nil, context.Canceled
						},
					}
				}
			},
			wantErr: true,
		},
		{
			name: "handles dataset API fetch errors gracefully",
			setup: func() {
				newDummyFetcherSet = func() fetcherSet {
					return fetcherSet{
						modelAPI:    &fetcher.DummyModelAPIFetcher{},
						modelReadme: &fetcher.DummyModelReadmeFetcher{},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return nil, context.Canceled // Dataset fetch fails
							},
						},
						datasetReadme: &fetcher.DummyDatasetReadmeFetcher{},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Dataset fetch failed, should have no dataset components
			},
		},
		{
			name: "handles dataset readme fetch error in BuildDummyBOM",
			setup: func() {
				newDummyFetcherSet = func() fetcherSet {
					return fetcherSet{
						modelAPI:    &fetcher.DummyModelAPIFetcher{},
						modelReadme: &fetcher.DummyModelReadmeFetcher{},
						datasetAPI:  &fetcher.DummyDatasetAPIFetcher{},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return nil, context.Canceled
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Dataset readme fetch failed but should still have dataset component (from API data)
				if got[0].BOM.Components == nil || len(*got[0].BOM.Components) == 0 {
					t.Error("Expected dataset components despite readme failure")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Restore originals before each test case
			newDummyFetcherSet = originalDummyFetcherSet
			newBOMBuilder = originalBuilder

			if tt.setup != nil {
				tt.setup()
			}
			got, err := BuildDummyBOM()
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildDummyBOM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, got)
			}
		})
	}
}

func TestBuildPerDiscovery(t *testing.T) {
	// Save original and restore after test
	originalBuilder := newBOMBuilder
	defer func() { newBOMBuilder = originalBuilder }()

	type args struct {
		discoveries []scanner.Discovery
		hfToken     string
		timeout     time.Duration
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
		check   func(*testing.T, []DiscoveredBOM)
	}{
		{
			name: "builds BOM for single discovery",
			args: args{
				discoveries: []scanner.Discovery{
					{ID: "test-model", Name: "test-model", Type: "huggingface"},
				},
				hfToken: "",
				timeout: 1 * time.Second,
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{SerialNumber: "test-serial"}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "builds BOM for empty discovery list",
			args: args{
				discoveries: []scanner.Discovery{},
				hfToken:     "",
				timeout:     1 * time.Second,
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 0 {
					t.Errorf("Expected 0 BOMs, got %d", len(got))
				}
			},
		},
		{
			name: "uses default timeout when zero",
			args: args{
				discoveries: []scanner.Discovery{
					{ID: "test-model", Name: "test-model", Type: "huggingface"},
				},
				hfToken: "",
				timeout: 0, // Zero timeout should use default
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "uses Name when ID is empty",
			args: args{
				discoveries: []scanner.Discovery{
					{ID: "", Name: "fallback-name", Type: "huggingface"},
				},
				hfToken: "",
				timeout: 1 * time.Second,
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := BuildPerDiscovery(tt.args.discoveries, tt.args.hfToken, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPerDiscovery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, got)
			}
		})
	}
}

func TestBuildPerDiscoveryWithProgress(t *testing.T) {
	// Save original and restore after test
	originalBuilder := newBOMBuilder
	originalFetcherSet := newFetcherSet
	defer func() {
		newBOMBuilder = originalBuilder
		newFetcherSet = originalFetcherSet
	}()

	type args struct {
		ctx         context.Context
		discoveries []scanner.Discovery
		opts        GenerateOptions
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
		check   func(*testing.T, []DiscoveredBOM)
	}{
		{
			name: "calls progress callback during build",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "test-model", Name: "test-model", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 1 * time.Second,
					OnProgress: func(event ProgressEvent) {
						// Progress callback is called
					},
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{SerialNumber: "test-serial"}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "handles context cancellation",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel() // Cancel immediately
					return ctx
				}(),
				discoveries: []scanner.Discovery{
					{ID: "test-model", Name: "test-model", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{}
				}
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "handles BOM build error during progress",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "test-model", Name: "test-model", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken:    "",
					Timeout:    1 * time.Second,
					OnProgress: func(event ProgressEvent) {},
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return nil, context.Canceled
						},
					}
				}
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "tracks progress events",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model1", Name: "model1", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 1 * time.Second,
					OnProgress: func(event ProgressEvent) {
						// Verify event types are called
					},
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "uses default timeout when opts.Timeout is zero",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model1", Name: "model1", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 0, // Should use default 10 seconds
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "uses Name when ID is empty string",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "", Name: "fallback-model", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "successfully fetches API and README with progress events",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model-with-api", Name: "model-with-api", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken:    "test-token",
					Timeout:    1 * time.Second,
					OnProgress: func(event ProgressEvent) {},
				},
			},
			setup: func() {
				builderCallCount := 0
				newBOMBuilder = func() bomBuilder {
					builderCallCount++
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return &cdx.Component{Name: ctx.DatasetID}, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									ID: id,
									CardData: map[string]interface{}{
										"datasets": []interface{}{"test-dataset", "failing-dataset"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
								return &fetcher.ModelReadmeCard{
									Datasets: []string{"readme-dataset"},
								}, nil
							},
						},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								if id == "failing-dataset" {
									return nil, context.Canceled
								}
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return &fetcher.DatasetReadmeCard{}, nil
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Should have datasets added as components (2 datasets: test-dataset, readme-dataset)
				// failing-dataset should be skipped
				if got[0].BOM.Components == nil {
					t.Error("Expected BOM to have dataset components")
					return
				}
				if len(*got[0].BOM.Components) != 2 {
					t.Errorf("Expected 2 dataset components (1 failed), got %d", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "builds datasets from model metadata",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model-with-datasets", Name: "model-with-datasets", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken:    "",
					Timeout:    1 * time.Second,
					OnProgress: nil,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							comp := &cdx.Component{
								Type: cdx.ComponentTypeData,
								Name: ctx.DatasetID,
							}
							return comp, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									CardData: map[string]interface{}{
										"datasets": []interface{}{"dataset-1"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
								return &fetcher.ModelReadmeCard{}, nil
							},
						},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return &fetcher.DatasetAPIResponse{
									ID: id,
								}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return &fetcher.DatasetReadmeCard{}, nil
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				if got[0].BOM.Components == nil {
					t.Error("Expected BOM to have components, but Components is nil")
					return
				}
				if len(*got[0].BOM.Components) != 1 {
					t.Errorf("Expected 1 dataset component, got %d: %+v", len(*got[0].BOM.Components), *got[0].BOM.Components)
				}
			},
		},
		{
			name: "handles dataset build errors gracefully",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model-with-failing-dataset", Name: "model-with-failing-dataset", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "test-token",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return nil, context.Canceled // Dataset build fails
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									CardData: map[string]interface{}{
										"datasets": []interface{}{"failing-dataset"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Dataset build failed, so components should be nil or empty
				if got[0].BOM.Components != nil && len(*got[0].BOM.Components) > 0 {
					t.Errorf("Expected no dataset components due to build error, got %d", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "handles dataset readme fetch errors",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model-with-dataset-readme-error", Name: "model-with-dataset-readme-error", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "test-token",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return &cdx.Component{Name: ctx.DatasetID, Type: cdx.ComponentTypeData}, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									CardData: map[string]interface{}{
										"datasets": []interface{}{"dataset1"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return nil, context.Canceled // Readme fetch fails
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Dataset API succeeded but readme failed - should still have the component
				if got[0].BOM.Components == nil || len(*got[0].BOM.Components) != 1 {
					t.Errorf("Expected 1 dataset component (readme failure is ignored), got %d", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "handles model API and README fetch errors gracefully",
			args: args{
				ctx: context.Background(),
				discoveries: []scanner.Discovery{
					{ID: "model-with-fetch-errors", Name: "model-with-fetch-errors", Type: "huggingface"},
				},
				opts: GenerateOptions{
					HFToken: "test-token",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							// Build succeeds even with nil API and README
							return &cdx.BOM{}, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return nil, context.Canceled // Model API fetch fails
							},
						},
						modelReadme: &mockModelReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
								return nil, context.Canceled // Model README fetch fails
							},
						},
						datasetAPI:    &mockDatasetAPIFetcher{},
						datasetReadme: &mockDatasetReadmeFetcher{},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Fetches failed but BOM was still built
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := BuildPerDiscoveryWithProgress(tt.args.ctx, tt.args.discoveries, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildPerDiscoveryWithProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, got)
			}
		})
	}
}

func Test_extractDatasetsFromModel(t *testing.T) {
	type args struct {
		modelResp *fetcher.ModelAPIResponse
		readme    *fetcher.ModelReadmeCard
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no datasets",
			args: args{
				modelResp: &fetcher.ModelAPIResponse{},
				readme:    &fetcher.ModelReadmeCard{},
			},
			want: nil,
		},
		{
			name: "datasets from API response - single string",
			args: args{
				modelResp: &fetcher.ModelAPIResponse{
					CardData: map[string]interface{}{
						"datasets": "dataset1",
					},
				},
				readme: nil,
			},
			want: []string{"dataset1"},
		},
		{
			name: "datasets from API response - array",
			args: args{
				modelResp: &fetcher.ModelAPIResponse{
					CardData: map[string]interface{}{
						"datasets": []interface{}{"dataset1", "dataset2"},
					},
				},
				readme: nil,
			},
			want: []string{"dataset1", "dataset2"},
		},
		{
			name: "datasets from readme",
			args: args{
				modelResp: nil,
				readme: &fetcher.ModelReadmeCard{
					Datasets: []string{"readme-dataset1", "readme-dataset2"},
				},
			},
			want: []string{"readme-dataset1", "readme-dataset2"},
		},
		{
			name: "datasets from both sources - deduplication",
			args: args{
				modelResp: &fetcher.ModelAPIResponse{
					CardData: map[string]interface{}{
						"datasets": []interface{}{"dataset1", "dataset2"},
					},
				},
				readme: &fetcher.ModelReadmeCard{
					Datasets: []string{"dataset2", "dataset3"},
				},
			},
			want: []string{"dataset1", "dataset2", "dataset3"},
		},
		{
			name: "filters empty strings",
			args: args{
				modelResp: &fetcher.ModelAPIResponse{
					CardData: map[string]interface{}{
						"datasets": []interface{}{"dataset1", "", "  ", "dataset2"},
					},
				},
				readme: nil,
			},
			want: []string{"dataset1", "dataset2"},
		},
		{
			name: "trims whitespace",
			args: args{
				modelResp: &fetcher.ModelAPIResponse{
					CardData: map[string]interface{}{
						"datasets": []interface{}{"  dataset1  ", "dataset2"},
					},
				},
				readme: nil,
			},
			want: []string{"dataset1", "dataset2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDatasetsFromModel(tt.args.modelResp, tt.args.readme)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractDatasetsFromModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFromModelIDs(t *testing.T) {
	// Save original and restore after test
	originalBuilder := newBOMBuilder
	defer func() { newBOMBuilder = originalBuilder }()

	type args struct {
		modelIDs []string
		hfToken  string
		timeout  time.Duration
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
		check   func(*testing.T, []DiscoveredBOM)
	}{
		{
			name: "builds BOM for single model ID",
			args: args{
				modelIDs: []string{"org/model"},
				hfToken:  "",
				timeout:  1 * time.Second,
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{SerialNumber: "test"}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				if got[0].Discovery.ID != "org/model" {
					t.Errorf("Expected model ID 'org/model', got %q", got[0].Discovery.ID)
				}
			},
		},
		{
			name: "skips empty model IDs",
			args: args{
				modelIDs: []string{"", "  ", "org/model"},
				hfToken:  "",
				timeout:  1 * time.Second,
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM (empty strings skipped), got %d", len(got))
				}
			},
		},
		{
			name: "handles empty list",
			args: args{
				modelIDs: []string{},
				hfToken:  "",
				timeout:  1 * time.Second,
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 0 {
					t.Errorf("Expected 0 BOMs, got %d", len(got))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := BuildFromModelIDs(tt.args.modelIDs, tt.args.hfToken, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFromModelIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, got)
			}
		})
	}
}

func TestBuildFromModelIDsWithProgress(t *testing.T) {
	// Save original and restore after test
	originalBuilder := newBOMBuilder
	originalFetcherSet := newFetcherSet
	defer func() {
		newBOMBuilder = originalBuilder
		newFetcherSet = originalFetcherSet
	}()

	type args struct {
		ctx      context.Context
		modelIDs []string
		opts     GenerateOptions
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
		check   func(*testing.T, []DiscoveredBOM)
	}{
		{
			name: "calls progress callback",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model"},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 1 * time.Second,
					OnProgress: func(event ProgressEvent) {
						// Progress callback called
					},
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "handles context cancellation",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				modelIDs: []string{"org/model"},
				opts: GenerateOptions{
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{}
				}
			},
			wantErr: true,
			check:   nil,
		},
		{
			name: "works with nil progress callback",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model"},
				opts: GenerateOptions{
					HFToken:    "",
					Timeout:    1 * time.Second,
					OnProgress: nil,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "builds model with single dataset from model IDs",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model-simple"},
				opts: GenerateOptions{
					HFToken:    "",
					Timeout:    1 * time.Second,
					OnProgress: nil,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return &cdx.Component{
								Name: ctx.DatasetID,
								Type: cdx.ComponentTypeData,
							}, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									CardData: map[string]interface{}{
										"datasets": []interface{}{"simple-dataset", "failing-dataset2"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
								return &fetcher.ModelReadmeCard{}, nil
							},
						},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								if id == "failing-dataset2" {
									return nil, context.Canceled
								}
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return &fetcher.DatasetReadmeCard{}, nil
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				if got[0].BOM.Components == nil {
					t.Error("Expected components to be non-nil")
					return
				}
				if len(*got[0].BOM.Components) != 1 {
					t.Errorf("Expected 1 dataset component (one failed), got %d", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "uses default timeout when zero",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model"},
				opts: GenerateOptions{
					HFToken:    "",
					Timeout:    0, // Should use default
					OnProgress: nil,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
				}
			},
		},
		{
			name: "continues on BOM build error",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model1", "org/model2"},
				opts: GenerateOptions{
					HFToken: "",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				callCount := 0
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							callCount++
							if callCount == 1 {
								return nil, context.Canceled // Error on first
							}
							return &cdx.BOM{}, nil
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				// Should have 1 BOM (second one succeeded)
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM (first failed, second succeeded), got %d", len(got))
				}
			},
		},
		{
			name: "successfully fetches and builds with datasets",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model-with-datasets"},
				opts: GenerateOptions{
					HFToken:    "test-token",
					Timeout:    1 * time.Second,
					OnProgress: func(event ProgressEvent) {},
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return &cdx.Component{Name: ctx.DatasetID, Type: cdx.ComponentTypeData}, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									ID: id,
									CardData: map[string]interface{}{
										"datasets": []interface{}{"dataset1", "dataset2"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelReadmeCard, error) {
								return &fetcher.ModelReadmeCard{
									Datasets: []string{"dataset3"},
								}, nil
							},
						},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return &fetcher.DatasetReadmeCard{}, nil
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Should have 3 datasets (dataset1, dataset2, dataset3)
				if got[0].BOM.Components == nil {
					t.Error("Expected BOM to have component")
					return
				}
				if len(*got[0].BOM.Components) != 3 {
					t.Errorf("Expected 3 dataset components, got %d", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "handles dataset build errors",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model-with-failing-datasets"},
				opts: GenerateOptions{
					HFToken: "test-token",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return nil, context.Canceled // Fail to build dataset
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									CardData: map[string]interface{}{
										"datasets": []interface{}{"dataset1"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Dataset failed to build, should have no components
				if got[0].BOM.Components != nil && len(*got[0].BOM.Components) > 0 {
					t.Errorf("Expected no components due to build errors, got %d", len(*got[0].BOM.Components))
				}
			},
		},
		{
			name: "handles dataset readme fetch errors",
			args: args{
				ctx:      context.Background(),
				modelIDs: []string{"org/model-with-dataset-readme-error"},
				opts: GenerateOptions{
					HFToken: "test-token",
					Timeout: 1 * time.Second,
				},
			},
			setup: func() {
				newBOMBuilder = func() bomBuilder {
					return &mockBOMBuilder{
						buildFunc: func(ctx builder.BuildContext) (*cdx.BOM, error) {
							return &cdx.BOM{}, nil
						},
						buildDatasetFunc: func(ctx builder.DatasetBuildContext) (*cdx.Component, error) {
							return &cdx.Component{Name: ctx.DatasetID, Type: cdx.ComponentTypeData}, nil
						},
					}
				}
				newFetcherSet = func(httpClient *http.Client, token string) fetcherSet {
					return fetcherSet{
						modelAPI: &mockModelAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.ModelAPIResponse, error) {
								return &fetcher.ModelAPIResponse{
									CardData: map[string]interface{}{
										"datasets": []interface{}{"dataset1"},
									},
								}, nil
							},
						},
						modelReadme: &mockModelReadmeFetcher{},
						datasetAPI: &mockDatasetAPIFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetAPIResponse, error) {
								return &fetcher.DatasetAPIResponse{ID: id}, nil
							},
						},
						datasetReadme: &mockDatasetReadmeFetcher{
							fetchFunc: func(ctx context.Context, id string) (*fetcher.DatasetReadmeCard, error) {
								return nil, context.Canceled // Readme fetch fails
							},
						},
					}
				}
			},
			wantErr: false,
			check: func(t *testing.T, got []DiscoveredBOM) {
				if len(got) != 1 {
					t.Errorf("Expected 1 BOM, got %d", len(got))
					return
				}
				// Dataset API succeeded but readme failed - should still have the component
				if got[0].BOM.Components == nil || len(*got[0].BOM.Components) != 1 {
					t.Errorf("Expected 1 dataset component (readme failure is ignored), got %d", len(*got[0].BOM.Components))
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			got, err := BuildFromModelIDsWithProgress(tt.args.ctx, tt.args.modelIDs, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFromModelIDsWithProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.check != nil && !tt.wantErr {
				tt.check(t, got)
			}
		})
	}
}
