//
//  Copyright 2026 The InfiniFlow Authors. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueritBuildsMinimalRequest(t *testing.T) {
	var gotMethod, gotPath, gotAuthorization, gotContentType string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		gotMethod = request.Method
		gotPath = request.URL.Path
		gotAuthorization = request.Header.Get("Authorization")
		gotContentType = request.Header.Get("Content-Type")
		_ = json.NewDecoder(request.Body).Decode(&gotBody)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"results":{"result":[]},"search_id":"search-1"}`))
	}))
	defer server.Close()

	querit := NewQueritToolWith(NewHTTPHelper().WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)}))
	out, err := querit.InvokableRun(context.Background(), `{"query":"ragflow","api_key":"key-test"}`)
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if gotMethod != http.MethodPost || gotPath != "/v1/search" {
		t.Fatalf("request = %s %s, want POST /v1/search", gotMethod, gotPath)
	}
	if gotAuthorization != "Bearer key-test" {
		t.Fatalf("Authorization = %q", gotAuthorization)
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Fatalf("Content-Type = %q", gotContentType)
	}
	if gotBody["query"] != "ragflow" || gotBody["count"] != float64(10) || gotBody["chunksPerDoc"] != float64(3) {
		t.Fatalf("request body = %#v", gotBody)
	}
	if _, exists := gotBody["filters"]; exists {
		t.Fatalf("empty filters must be omitted: %#v", gotBody)
	}
	if !strings.Contains(out, `"search_id":"search-1"`) {
		t.Fatalf("complete response was not retained: %s", out)
	}
}

func TestQueritBuildsFiltersAndMergesRuntimeOverrides(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewDecoder(request.Body).Decode(&gotBody)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"results":{"result":[]}}`))
	}))
	defer server.Close()

	defaults := queritParams{
		APIKey:          "stored-key",
		Count:           20,
		ChunksPerDoc:    2,
		SiteInclude:     []string{"stored.example"},
		SiteExclude:     []string{"blocked.example"},
		TimeRange:       "w1",
		CountryInclude:  []string{"CN"},
		LanguageInclude: []string{"zh"},
	}
	helper := NewHTTPHelper().WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)})
	querit := newQueritTool(helper, func() string { return "" }, defaults, func(context.Context, int) bool { return true })
	_, err := querit.InvokableRun(context.Background(), `{"query":"ragflow","count":5,"site_include":[],"language_include":["en"]}`)
	if err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if gotBody["count"] != float64(5) || gotBody["chunksPerDoc"] != float64(2) {
		t.Fatalf("merged scalar defaults = %#v", gotBody)
	}
	filters := gotBody["filters"].(map[string]any)
	sites := filters["sites"].(map[string]any)
	if _, exists := sites["include"]; exists || len(sites["exclude"].([]any)) != 1 {
		t.Fatalf("explicit empty site_include did not clear node default: %#v", sites)
	}
	if filters["timeRange"].(map[string]any)["date"] != "w1" {
		t.Fatalf("timeRange = %#v", filters["timeRange"])
	}
	if filters["geo"].(map[string]any)["countries"].(map[string]any)["include"].([]any)[0] != "CN" {
		t.Fatalf("geo filter = %#v", filters["geo"])
	}
	if filters["languages"].(map[string]any)["include"].([]any)[0] != "en" {
		t.Fatalf("language filter = %#v", filters["languages"])
	}
}

func TestQueritAPIKeyResolutionAndEmptyQuery(t *testing.T) {
	var calls atomic.Int32
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		authorization = request.Header.Get("Authorization")
		_, _ = writer.Write([]byte(`{"results":{"result":[]}}`))
	}))
	defer server.Close()

	helper := NewHTTPHelper().WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)})
	querit := NewQueritToolWithEnvKey(helper, func() string { return "environment-secret" })
	if _, err := querit.InvokableRun(context.Background(), `{"query":"ragflow"}`); err != nil {
		t.Fatalf("InvokableRun: %v", err)
	}
	if authorization != "Bearer environment-secret" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if _, err := querit.InvokableRun(context.Background(), `{"query":""}`); err != nil {
		t.Fatalf("empty query: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("empty query made a request; calls = %d", calls.Load())
	}

	missing := NewQueritToolWithEnvKey(helper, func() string { return "" })
	out, err := missing.InvokableRun(context.Background(), `{"query":"ragflow"}`)
	if err != nil {
		t.Fatalf("missing key returned Go error: %v", err)
	}
	if !strings.Contains(out, "api_key") || strings.Contains(out, "environment-secret") || calls.Load() != 1 {
		t.Fatalf("missing-key result = %s, calls = %d", out, calls.Load())
	}
}

func TestQueritValidatesParametersBeforeRequest(t *testing.T) {
	tests := []string{
		`{"query":"x","api_key":"k","count":0}`,
		`{"query":"x","api_key":"k","chunks_per_doc":4}`,
		`{"query":"x","api_key":"k","time_range":"last week"}`,
		`{"query":"x","api_key":"k","time_range":"2026-01-01,2026-01-31"}`,
		`{"query":"x","api_key":"k","site_include":[1]}`,
	}
	for _, args := range tests {
		t.Run(args, func(t *testing.T) {
			out, err := NewQueritTool().InvokableRun(context.Background(), args)
			if err != nil || !strings.Contains(out, "_ERROR") {
				t.Fatalf("InvokableRun(%s) = %s, %v", args, out, err)
			}
		})
	}
}

func TestQueritTimeRangeContract(t *testing.T) {
	tests := []struct {
		value string
		valid bool
	}{
		{value: "", valid: true},
		{value: "d7", valid: true},
		{value: "w2", valid: true},
		{value: "m3", valid: true},
		{value: "y1", valid: true},
		{value: "2026-01-01to2026-01-31", valid: true},
		{value: "d0", valid: false},
		{value: "7d", valid: false},
		{value: "2026-01-01,2026-01-31", valid: false},
		{value: "2026-01-01..2026-01-31", valid: false},
		{value: "2026-01-01-2026-01-31", valid: false},
		{value: "last week", valid: false},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			if got := isValidQueritTimeRange(test.value); got != test.valid {
				t.Fatalf("isValidQueritTimeRange(%q) = %v, want %v", test.value, got, test.valid)
			}
			if !test.valid || test.value == "" {
				return
			}
			request := buildQueritRequest(queritParams{TimeRange: test.value})
			if request.Filters == nil || request.Filters.TimeRange == nil || request.Filters.TimeRange.Date != test.value {
				t.Fatalf("time range request mapping = %#v", request.Filters)
			}
		})
	}
}

func TestQueritHTTPFailuresAreSoftErrors(t *testing.T) {
	t.Run("unauthorized is not retried", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			writer.WriteHeader(http.StatusUnauthorized)
			_, _ = writer.Write([]byte(`{"message":"do not expose upstream bodies"}`))
		}))
		defer server.Close()
		querit := NewQueritToolWith(NewHTTPHelper().WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)}))
		out, err := querit.InvokableRun(context.Background(), `{"query":"x","api_key":"secret-key"}`)
		if err != nil || calls.Load() != 1 || !strings.Contains(out, "401") || strings.Contains(out, "secret-key") {
			t.Fatalf("result = %s, err = %v, calls = %d", out, err, calls.Load())
		}
	})

	t.Run("rate limit is retried at most three times", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			writer.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()
		helper := NewHTTPHelper().WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)})
		querit := newQueritTool(helper, nil, queritParams{}, func(context.Context, int) bool { return true })
		out, err := querit.InvokableRun(context.Background(), `{"query":"x","api_key":"secret-key"}`)
		if err != nil || calls.Load() != queritMaxAttempts || !strings.Contains(out, "429") || strings.Contains(out, "secret-key") {
			t.Fatalf("result = %s, err = %v, calls = %d", out, err, calls.Load())
		}
	})

	t.Run("HTTP helper retries temporary server errors", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			attempt := calls.Add(1)
			if attempt < 3 {
				writer.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = writer.Write([]byte(`{"results":{"result":[]}}`))
		}))
		defer server.Close()
		helper := NewHTTPHelperWithRetry(RetryConfig{
			MaxAttempts: 3,
			BaseBackoff: time.Nanosecond,
			MaxBackoff:  time.Nanosecond,
		}).WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)})
		out, err := NewQueritToolWith(helper).InvokableRun(context.Background(), `{"query":"x","api_key":"k"}`)
		if err != nil || calls.Load() != 3 || strings.Contains(out, "_ERROR") {
			t.Fatalf("result = %s, err = %v, calls = %d", out, err, calls.Load())
		}
	})

	t.Run("persistent server errors are soft and redact the environment key", func(t *testing.T) {
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			writer.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()
		helper := NewHTTPHelperWithRetry(RetryConfig{
			MaxAttempts: 3,
			BaseBackoff: time.Nanosecond,
			MaxBackoff:  time.Nanosecond,
		}).WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)})
		querit := NewQueritToolWithEnvKey(helper, func() string { return "environment-secret" })
		out, err := querit.InvokableRun(context.Background(), `{"query":"x"}`)
		if err != nil || calls.Load() != 3 || !strings.Contains(out, "_ERROR") || !strings.Contains(out, "500") {
			t.Fatalf("result = %s, err = %v, calls = %d", out, err, calls.Load())
		}
		if strings.Contains(out, "environment-secret") {
			t.Fatalf("soft error exposed environment API key: %s", out)
		}
	})

	t.Run("network errors are soft", func(t *testing.T) {
		var calls atomic.Int32
		helper := NewHTTPHelperWithRetry(RetryConfig{
			MaxAttempts: 3,
			BaseBackoff: time.Nanosecond,
			MaxBackoff:  time.Nanosecond,
		}).WithClient(&http.Client{Transport: roundTripperErrorFunc(func(*http.Request) error {
			calls.Add(1)
			return errors.New("offline")
		})})
		out, err := NewQueritToolWith(helper).InvokableRun(context.Background(), `{"query":"x","api_key":"k"}`)
		if err != nil || calls.Load() != 3 || !strings.Contains(out, "_ERROR") {
			t.Fatalf("result = %s, err = %v, calls = %d", out, err, calls.Load())
		}
	})
}

func TestQueritRejectsInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write([]byte(`not-json`))
	}))
	defer server.Close()
	querit := NewQueritToolWith(NewHTTPHelper().WithClient(&http.Client{Transport: rewriteHostTransport(server.URL)}))
	out, err := querit.InvokableRun(context.Background(), `{"query":"x","api_key":"k"}`)
	if err != nil || !strings.Contains(out, "decode response") {
		t.Fatalf("result = %s, err = %v", out, err)
	}
}

func TestQueritReferencesAndCompleteComponentOutput(t *testing.T) {
	response := map[string]any{
		"search_id":     "search-1",
		"query_context": map[string]any{"rewritten": "rag flow"},
		"results": map[string]any{"result": []any{
			map[string]any{"title": "RAGFlow", "url": "https://ragflow.io", "snippet": "RAG engine", "extra": true},
			map[string]any{"title": "Missing optional values"},
			"invalid item",
		}},
	}
	querit := NewQueritTool()
	chunks, docAggs := querit.BuildReferences(context.Background(), response)
	if len(chunks) != 2 || len(docAggs) != 2 {
		t.Fatalf("references = %#v / %#v", chunks, docAggs)
	}
	if chunks[0]["content"] != "RAG engine" || chunks[0]["score"] != 1 || chunks[0]["similarity"] != 1 {
		t.Fatalf("reference = %#v", chunks[0])
	}
	outputs := querit.BuildComponentOutputs(response)
	if outputs["json"].(map[string]any)["search_id"] != "search-1" {
		t.Fatalf("complete json output = %#v", outputs["json"])
	}
	formatted := outputs["formalized_content"].(string)
	for _, expected := range []string{"Title: RAGFlow", "URL: https://ragflow.io", "RAG engine"} {
		if !strings.Contains(formatted, expected) {
			t.Fatalf("formalized_content missing %q: %s", expected, formatted)
		}
	}
	if chunks, docAggs := querit.BuildReferences(context.Background(), map[string]any{"results": nil}); len(chunks) != 0 || len(docAggs) != 0 {
		t.Fatalf("malformed response references = %#v / %#v", chunks, docAggs)
	}
}

func TestQueritInfoDoesNotExposeAPIKey(t *testing.T) {
	info, err := NewQueritTool().Info(context.Background())
	if err != nil {
		t.Fatalf("Info: %v", err)
	}
	if info.Name != queritToolName || info.Desc == "" || info.ParamsOneOf == nil {
		t.Fatalf("Info = %#v", info)
	}
	encoded, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal Info: %v", err)
	}
	if strings.Contains(string(encoded), "api_key") {
		t.Fatalf("Info exposed API key: %s", encoded)
	}
}

type roundTripperErrorFunc func(*http.Request) error

func (f roundTripperErrorFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return nil, f(request)
}
