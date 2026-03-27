from __future__ import annotations

from app.flask_app import create_app


def test_root_route() -> None:
    app = create_app()
    client = app.test_client()

    response = client.get("/")

    assert response.status_code == 200
    assert response.get_json() == {"message": "Hello World"}


def test_greet_route_default_from() -> None:
    app = create_app()
    client = app.test_client()

    response = client.get("/api/greet/jarry")

    assert response.status_code == 200
    assert response.get_json() == {"message": "Hello, jarry!", "from": "starter"}


def test_greet_route_custom_from() -> None:
    app = create_app()
    client = app.test_client()

    response = client.get("/api/greet/jarry?from=local")

    assert response.status_code == 200
    assert response.get_json() == {"message": "Hello, jarry!", "from": "local"}
