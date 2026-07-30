#
#  Copyright 2026 The InfiniFlow Authors. All Rights Reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

from typing import Protocol

from rag.utils.querit_conn import Querit
from rag.utils.tavily_conn import Tavily

WEB_SEARCH_PROVIDER_TAVILY = "tavily"
WEB_SEARCH_PROVIDER_QUERIT = "querit"


class WebSearchProvider(Protocol):
    def retrieve_chunks(self, question: str) -> dict[str, list]:
        """Return web results in RAGFlow's chunk and document aggregate shape."""


def has_web_search_provider(prompt_config: dict | None) -> bool:
    if not prompt_config:
        return False
    provider = prompt_config.get("web_search_provider", WEB_SEARCH_PROVIDER_TAVILY)
    if provider == WEB_SEARCH_PROVIDER_TAVILY:
        return bool(prompt_config.get("tavily_api_key"))
    if provider == WEB_SEARCH_PROVIDER_QUERIT:
        return bool(prompt_config.get("querit_api_key"))
    return False


def create_web_search_provider(prompt_config: dict | None) -> WebSearchProvider | None:
    if not has_web_search_provider(prompt_config):
        return None
    if prompt_config.get("web_search_provider", WEB_SEARCH_PROVIDER_TAVILY) == WEB_SEARCH_PROVIDER_QUERIT:
        return Querit(prompt_config["querit_api_key"])
    return Tavily(prompt_config["tavily_api_key"])
