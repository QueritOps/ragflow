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

from rag.utils.tavily_conn import Tavily


class WebSearchProvider(Protocol):
    def retrieve_chunks(self, question: str) -> dict[str, list]:
        """Return web results in RAGFlow's chunk and document aggregate shape."""


def has_web_search_provider(prompt_config: dict | None) -> bool:
    return bool(prompt_config and prompt_config.get("tavily_api_key"))


def create_web_search_provider(prompt_config: dict | None) -> WebSearchProvider | None:
    if not has_web_search_provider(prompt_config):
        return None
    return Tavily(prompt_config["tavily_api_key"])
