from __future__ import annotations

import base64
from collections.abc import Mapping
from typing import Any
from urllib.parse import parse_qsl

from aws_lambda_typing import events as aws_events

from app.events import is_alb, is_http_v1, is_http_v2, is_sqs
from app.flask_app import create_app


def handler(event: Any, _context: Any) -> Any:
    if is_http_v2(event):
        return _convert_http_event(event, "v2")
    if is_alb(event):
        return _convert_http_event(event, "v1")
    if is_http_v1(event):
        return _convert_http_event(event, "v1")
    if is_sqs(event):
        return _handle_sqs(event)
    raise ValueError("Unsupported event type.")


def _handle_sqs(event: aws_events.SQSEvent) -> None:
    records = event.get("Records", [])
    for record in records:
        if isinstance(record, dict):
            message_id = record.get("messageId", "unknown")
            print(f"SQS message received: {message_id}")
    return None


def _convert_http_event(
    event: aws_events.APIGatewayProxyEventV2 | aws_events.APIGatewayProxyEventV1 | aws_events.ALBEvent, version: str
) -> dict[str, Any]:
    app = create_app()
    client = app.test_client()

    if version == "v2":
        request_context = event.get("requestContext", {})
        http = request_context.get("http", {}) if isinstance(request_context, dict) else {}
        method = str(http.get("method", "GET")).upper()
        path = str(event.get("rawPath") or http.get("path") or "/")
        query = _query_for_v2(event)
    else:
        method = str(event.get("httpMethod", "GET")).upper()
        path = str(event.get("path") or "/")
        query = _query_for_v1(event)

    headers = _normalize_headers(event.get("headers"))
    body = _event_body(event)
    response = client.open(path=path, method=method, query_string=query, headers=headers, data=body)

    return {
        "statusCode": response.status_code,
        "headers": dict(response.headers),
        "body": response.get_data(as_text=True),
    }


def _query_for_v2(event: Mapping[str, Any]) -> list[tuple[str, str]]:
    raw_query = event.get("rawQueryString")
    if isinstance(raw_query, str) and raw_query.strip():
        return [(k, v) for k, v in parse_qsl(raw_query, keep_blank_values=True)]
    return _query_for_v1(event)


def _query_for_v1(event: Mapping[str, Any]) -> list[tuple[str, str]]:
    query = event.get("queryStringParameters")
    if not isinstance(query, dict):
        return []
    pairs: list[tuple[str, str]] = []
    for key, value in query.items():
        if key is None or value is None:
            continue
        pairs.append((str(key), str(value)))
    return pairs


def _normalize_headers(raw_headers: Any) -> dict[str, str]:
    if not isinstance(raw_headers, dict):
        return {}
    headers: dict[str, str] = {}
    for key, value in raw_headers.items():
        if key is None or value is None:
            continue
        headers[str(key)] = str(value)
    return headers


def _event_body(event: Mapping[str, Any]) -> bytes | None:
    raw_body = event.get("body")
    if not isinstance(raw_body, str):
        return None
    if event.get("isBase64Encoded") is True:
        return base64.b64decode(raw_body)
    return raw_body.encode("utf-8")
