from __future__ import annotations

import argparse
import json
import os
from pathlib import Path
from typing import Any

from app.handler import handler
from app.flask_app import create_app


def main() -> None:
    parser = argparse.ArgumentParser(description="AWS Lambda Python starter CLI")
    subparsers = parser.add_subparsers(dest="command")

    dev_cmd = subparsers.add_parser("dev", help="Start local Flask server")
    dev_cmd.add_argument("--port", "-p", type=int, help="Local HTTP port")

    test_event_cmd = subparsers.add_parser("test-event", help="Invoke handler with local event JSON")
    test_event_cmd.add_argument("event_json_file", help="Path to event JSON file")

    subparsers.add_parser("lambda", help="Alias to lambda runtime entrypoint")

    args = parser.parse_args()
    command = args.command or "lambda"

    if command == "dev":
        run_dev(port=args.port)
        return
    if command == "test-event":
        run_test_event(args.event_json_file)
        return
    if command == "lambda":
        os.environ.setdefault("_HANDLER", "app.entrypoints.lambda_handler.handler")
        from awslambdaric.__main__ import main as ric_main

        ric_main()
        return
    parser.error(f"Unsupported command: {command}")


def run_dev(port: int | None = None) -> None:
    env_port = os.getenv("PORT", "").strip()
    default_port = 3000
    if env_port.isdigit() and int(env_port) > 0:
        default_port = int(env_port)
    final_port = port if isinstance(port, int) and port > 0 else default_port
    app = create_app()
    app.run(host="0.0.0.0", port=final_port)


def run_test_event(file_path: str) -> None:
    payload = load_event_file(file_path)
    result = handler(payload, None)
    if result is None:
        print("Handler completed with no return value.")
        return
    print(json.dumps(result, ensure_ascii=False, indent=2))


def load_event_file(file_path: str) -> Any:
    path = Path(file_path)
    if not path.exists():
        raise FileNotFoundError(f'failed to read event file "{file_path}": file does not exist')
    content = path.read_text(encoding="utf-8")
    if content == "":
        raise ValueError("event file is empty")
    return json.loads(content)


if __name__ == "__main__":
    main()
