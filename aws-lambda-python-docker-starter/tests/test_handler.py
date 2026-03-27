from __future__ import annotations

import json
from pathlib import Path

import pytest

from app.handler import handler


def _load_fixture(name: str) -> dict:
    path = Path(__file__).resolve().parents[1] / "test-cases" / name
    return json.loads(path.read_text(encoding="utf-8"))


def test_handler_http_v2_fixture() -> None:
    event = _load_fixture("http-v2-event.json")

    result = handler(event, None)

    assert result["statusCode"] == 200
    assert "application/json" in result["headers"]["Content-Type"]
    assert json.loads(result["body"]) == {"message": "Hello, jarry!", "from": "local-v2"}


def test_handler_http_v1_fixture() -> None:
    event = _load_fixture("http-v1-event.json")

    result = handler(event, None)

    assert result["statusCode"] == 200
    assert "application/json" in result["headers"]["Content-Type"]
    assert json.loads(result["body"]) == {"message": "Hello, jarry!", "from": "local-v1"}


def test_handler_alb_event() -> None:
    event = {
        "httpMethod": "GET",
        "path": "/api/greet/jarry",
        "queryStringParameters": {"from": "alb"},
        "requestContext": {"elb": {"targetGroupArn": "arn:aws:elasticloadbalancing:example"}},
    }

    result = handler(event, None)

    assert result["statusCode"] == 200
    assert "application/json" in result["headers"]["Content-Type"]
    assert json.loads(result["body"]) == {"message": "Hello, jarry!", "from": "alb"}


def test_handler_sqs_fixture() -> None:
    event = _load_fixture("sqs-event.json")

    result = handler(event, None)

    assert result is None


def test_handler_unsupported_event() -> None:
    with pytest.raises(ValueError, match="Unsupported event type."):
        handler({"foo": "bar"}, None)
