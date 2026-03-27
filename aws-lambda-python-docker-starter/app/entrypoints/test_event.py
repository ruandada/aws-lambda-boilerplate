from __future__ import annotations

import argparse

from app.cli import run_test_event


def main() -> None:
    parser = argparse.ArgumentParser(description="Invoke Lambda handler with local event file")
    parser.add_argument("event_json_file")
    args = parser.parse_args()
    run_test_event(args.event_json_file)


if __name__ == "__main__":
    main()
