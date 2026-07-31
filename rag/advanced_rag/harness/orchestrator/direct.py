"""Low mode: direct single-pass search."""

import asyncio
import logging

from rag.advanced_rag.harness.tools.search import hybrid_search, web_search

_LOG = logging.getLogger(__name__)


async def direct_search(state: dict, tools) -> dict:
    """Search each configured evidence source once and merge into kbinfos."""
    question = state.get("question", "")
    keywords = state.get("keywords", "")
    _LOG.info('[Direct search] Looking up configured evidence sources for: "%s" (keywords: %s)', question, keywords)

    searches = [hybrid_search(tools, query=question, keywords=keywords)]
    if tools.has_web():
        searches.append(web_search(tools, query=question, keywords=keywords))

    for result in await asyncio.gather(*searches):
        _merge_kbinfos(tools, result)

    if not _has_chunks(tools):
        _LOG.info("[Direct search] Found no matching passages.")
        return {"empty_result": True, "kbinfos": tools.kbinfos}

    return {"kbinfos": tools.kbinfos}


def _merge_kbinfos(tools, result: dict):
    if not result or not result.get("chunks"):
        return
    seen = {_chunk_key(c) for c in tools.kbinfos.get("chunks", [])}
    for c in result.get("chunks", []):
        k = _chunk_key(c)
        if k in seen:
            continue
        seen.add(k)
        tools.kbinfos.setdefault("chunks", []).append(c)
    dseen = {d.get("doc_id") for d in tools.kbinfos.get("doc_aggs", [])}
    for d in result.get("doc_aggs", []):
        if d.get("doc_id") in dseen:
            continue
        dseen.add(d.get("doc_id"))
        tools.kbinfos.setdefault("doc_aggs", []).append(d)


def _chunk_key(ck: dict) -> str:
    return ck.get("chunk_id") or ck.get("id") or str(id(ck))


def _has_chunks(tools) -> bool:
    return bool(tools.kbinfos.get("chunks"))
