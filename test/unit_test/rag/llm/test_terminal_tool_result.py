#
#  Copyright 2026 The InfiniFlow Authors. All Rights Reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#

from rag.llm.chat_model import _format_reasoning_delta, _terminal_tool_result


def test_terminal_tool_result_returns_successful_final_answer():
    results = [(object(), "rag", {}, "**ERROR**: Web search failed.", None)]

    assert _terminal_tool_result(results, {"rag"}) == (True, "**ERROR**: Web search failed.")


def test_terminal_tool_result_serializes_non_string_results():
    results = [(object(), "rag", {}, {"answer": "done"}, None)]

    assert _terminal_tool_result(results, {"rag"}) == (True, '{"answer": "done"}')


def test_terminal_tool_result_ignores_errors_and_non_terminal_tools():
    failed = [(object(), "rag", {}, None, RuntimeError("failed"))]
    other = [(object(), "summarize_document", {}, "summary", None)]

    assert _terminal_tool_result(failed, {"rag"}) == (False, "")
    assert _terminal_tool_result(other, {"rag"}) == (False, "")


def test_reasoning_deltas_open_once_and_close_on_visible_content():
    first, in_reasoning = _format_reasoning_delta("step one", "", False)
    second, in_reasoning = _format_reasoning_delta(" step two", "", in_reasoning)
    visible, in_reasoning = _format_reasoning_delta(None, "final answer", in_reasoning)

    assert first == "<think>step one"
    assert second == " step two"
    assert visible == "</think>final answer"
    assert in_reasoning is False


def test_reasoning_delta_can_hide_reasoning_without_leaking_tags():
    text, in_reasoning = _format_reasoning_delta("private reasoning", "", False, include_reasoning=False)

    assert text == ""
    assert in_reasoning is False
