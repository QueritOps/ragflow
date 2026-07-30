import { WebSearchProvider } from '@/constants/chat';
import { useFetchChat } from '@/hooks/use-chat-request';
import type { PromptConfig } from '@/interfaces/database/chat';
import { isEmpty } from 'lodash';

export function getWebSearchApiKey(promptConfig?: PromptConfig) {
  const provider =
    promptConfig?.web_search_provider ?? WebSearchProvider.Tavily;
  const apiKey =
    provider === WebSearchProvider.Querit
      ? promptConfig?.querit_api_key
      : promptConfig?.tavily_api_key;
  return apiKey?.trim();
}

export function useShowInternet() {
  const { data: currentDialog } = useFetchChat();

  return !isEmpty(getWebSearchApiKey(currentDialog?.prompt_config));
}
