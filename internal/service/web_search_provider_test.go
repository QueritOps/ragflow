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

package service

import "testing"

func TestResolveWebSearchProviderUsesExistingTavilyConfig(t *testing.T) {
	provider := resolveWebSearchProvider(map[string]interface{}{
		"tavily_api_key": "tvly-test",
	})

	if provider == nil {
		t.Fatal("provider is nil")
	}
	if provider.Provider != webSearchProviderTavily {
		t.Fatalf("provider = %q, want %q", provider.Provider, webSearchProviderTavily)
	}
	if provider.APIKey != "tvly-test" {
		t.Fatalf("api key = %q, want %q", provider.APIKey, "tvly-test")
	}
}

func TestResolveWebSearchProviderReturnsNilWithoutTavilyKey(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]interface{}
	}{
		{name: "nil config", config: nil},
		{name: "empty config", config: map[string]interface{}{}},
		{name: "empty key", config: map[string]interface{}{"tavily_api_key": ""}},
		{name: "non-string key", config: map[string]interface{}{"tavily_api_key": 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if provider := resolveWebSearchProvider(tc.config); provider != nil {
				t.Fatalf("provider = %+v, want nil", provider)
			}
		})
	}
}
