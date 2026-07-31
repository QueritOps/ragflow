#
#  Copyright 2026 The InfiniFlow Authors. All Rights Reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#

WEB_SEARCH_FAILURE_MESSAGE = "**ERROR**: Web search failed. Check the selected provider API Key and try again."


class WebSearchProviderError(RuntimeError):
    """Raised when a configured web-search provider cannot complete a request."""
