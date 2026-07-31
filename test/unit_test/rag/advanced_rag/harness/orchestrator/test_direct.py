from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest

from rag.advanced_rag.harness.orchestrator import direct


def _tools(*, has_web: bool):
    return SimpleNamespace(
        has_web=lambda: has_web,
        kbinfos={"chunks": [], "doc_aggs": []},
    )


@pytest.mark.asyncio
async def test_direct_search_uses_web_when_configured_without_knowledge_base(monkeypatch):
    tools = _tools(has_web=True)
    hybrid = AsyncMock(return_value={"chunks": [], "doc_aggs": []})
    web = AsyncMock(
        return_value={
            "chunks": [{"chunk_id": "web-1", "doc_id": "web-doc-1", "content_with_weight": "live evidence"}],
            "doc_aggs": [{"doc_id": "web-doc-1", "doc_name": "Live source"}],
        }
    )
    monkeypatch.setattr(direct, "hybrid_search", hybrid)
    monkeypatch.setattr(direct, "web_search", web)

    result = await direct.direct_search({"question": "latest news", "keywords": "news"}, tools)

    hybrid.assert_awaited_once_with(tools, query="latest news", keywords="news")
    web.assert_awaited_once_with(tools, query="latest news", keywords="news")
    assert result["kbinfos"]["chunks"][0]["chunk_id"] == "web-1"
    assert "empty_result" not in result


@pytest.mark.asyncio
async def test_direct_search_skips_web_when_not_configured(monkeypatch):
    tools = _tools(has_web=False)
    hybrid = AsyncMock(return_value={"chunks": [], "doc_aggs": []})
    web = AsyncMock()
    monkeypatch.setattr(direct, "hybrid_search", hybrid)
    monkeypatch.setattr(direct, "web_search", web)

    result = await direct.direct_search({"question": "local question"}, tools)

    web.assert_not_awaited()
    assert result["empty_result"] is True


@pytest.mark.asyncio
async def test_direct_search_merges_and_deduplicates_knowledge_and_web(monkeypatch):
    tools = _tools(has_web=True)
    hybrid = AsyncMock(
        return_value={
            "chunks": [{"chunk_id": "shared", "doc_id": "kb-doc"}, {"chunk_id": "kb-2", "doc_id": "kb-doc-2"}],
            "doc_aggs": [{"doc_id": "kb-doc"}, {"doc_id": "kb-doc-2"}],
        }
    )
    web = AsyncMock(
        return_value={
            "chunks": [{"chunk_id": "shared", "doc_id": "web-duplicate"}, {"chunk_id": "web-2", "doc_id": "web-doc-2"}],
            "doc_aggs": [{"doc_id": "kb-doc"}, {"doc_id": "web-doc-2"}],
        }
    )
    monkeypatch.setattr(direct, "hybrid_search", hybrid)
    monkeypatch.setattr(direct, "web_search", web)

    result = await direct.direct_search({"question": "combined"}, tools)

    assert [chunk["chunk_id"] for chunk in result["kbinfos"]["chunks"]] == ["shared", "kb-2", "web-2"]
    assert [doc["doc_id"] for doc in result["kbinfos"]["doc_aggs"]] == ["kb-doc", "kb-doc-2", "web-doc-2"]
