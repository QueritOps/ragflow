import { WebSearchProvider } from '@/constants/chat';
import type { PromptConfig } from '@/interfaces/database/chat';

export function getWebSearchApiKey(promptConfig?: PromptConfig) {
  const provider =
    promptConfig?.web_search_provider ?? WebSearchProvider.Tavily;
  let apiKey: unknown;

  switch (provider) {
    case WebSearchProvider.Tavily:
      apiKey = promptConfig?.tavily_api_key;
      break;
    case WebSearchProvider.Querit:
      apiKey = promptConfig?.querit_api_key;
      break;
    default:
      return undefined;
  }

  return typeof apiKey === 'string' ? apiKey.trim() : undefined;
}
