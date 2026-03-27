from __future__ import annotations

from flask import Flask, jsonify, request


def create_app() -> Flask:
    app = Flask(__name__)

    @app.get("/")
    def root() -> tuple[dict[str, str], int]:
        return {"message": "Hello World"}, 200

    @app.get("/api/greet/<name>")
    def greet(name: str) -> tuple[dict[str, str], int]:
        from_value = request.args.get("from", "").strip() or "starter"
        return jsonify({"message": f"Hello, {name}!", "from": from_value}), 200

    return app
