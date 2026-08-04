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

import (
	"context"
	"fmt"
)

const webSearchProviderTavily = "tavily"

type webSearchProviderConfig struct {
	Provider string
	APIKey   string
}

func resolveWebSearchProvider(promptConfig map[string]interface{}) *webSearchProviderConfig {
	if promptConfig == nil {
		return nil
	}
	apiKey, _ := promptConfig["tavily_api_key"].(string)
	if apiKey == "" {
		return nil
	}
	return &webSearchProviderConfig{
		Provider: webSearchProviderTavily,
		APIKey:   apiKey,
	}
}

func (s *ChatPipelineService) retrieveWebSearch(
	ctx context.Context,
	provider *webSearchProviderConfig,
	question string,
) (map[string]interface{}, error) {
	if provider == nil {
		return nil, fmt.Errorf("web search provider is not configured")
	}
	switch provider.Provider {
	case webSearchProviderTavily:
		return s.tavilyRetrieve(ctx, provider.APIKey, question)
	default:
		return nil, fmt.Errorf("unsupported web search provider %q", provider.Provider)
	}
}

func (dr *DeepResearcher) retrieveWebSearch(
	ctx context.Context,
	provider *webSearchProviderConfig,
	query string,
) (map[string]interface{}, error) {
	if provider == nil {
		return nil, fmt.Errorf("web search provider is not configured")
	}
	switch provider.Provider {
	case webSearchProviderTavily:
		return dr.tavilyRetrieve(ctx, provider.APIKey, query)
	default:
		return nil, fmt.Errorf("unsupported web search provider %q", provider.Provider)
	}
}
