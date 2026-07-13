#!/usr/bin/env python3
"""Prepare or validate the semantic version carried by a pull request."""

from __future__ import annotations

import argparse
import json
import re
from datetime import date
from pathlib import Path

SEMVER = re.compile(r"^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$")
CONVENTIONAL = re.compile(
    r"^(?P<type>[a-z][a-z0-9-]*)(?:\([^)]+\))?(?P<breaking>!)?:\s*(?P<summary>.+)$",
    re.IGNORECASE,
)


def parse_version(value: str) -> tuple[int, int, int]:
    match = SEMVER.fullmatch(value.strip())
    if not match:
        raise ValueError(f"invalid semantic version: {value!r}")
    return tuple(int(part) for part in match.groups())  # type: ignore[return-value]


def classify(title: str, body: str) -> tuple[str, bool, str]:
    match = CONVENTIONAL.match(title.strip())
    if not match:
        return "none", "BREAKING CHANGE:" in body, title.strip()

    change_type = match.group("type").lower()
    breaking = bool(match.group("breaking")) or "BREAKING CHANGE:" in body
    return change_type, breaking, match.group("summary").strip()


def next_version(base: str, title: str, body: str) -> str:
    major, minor, patch = parse_version(base)
    change_type, breaking, _ = classify(title, body)

    if breaking:
        if major == 0:
            return f"0.{minor + 1}.0"
        return f"{major + 1}.0.0"
    if change_type == "feat":
        return f"{major}.{minor + 1}.0"
    if change_type == "fix":
        return f"{major}.{minor}.{patch + 1}"
    return f"{major}.{minor}.{patch}"


def changelog_section(title: str, body: str) -> tuple[str, str]:
    change_type, breaking, summary = classify(title, body)
    heading = "Changed" if breaking else "Fixed" if change_type == "fix" else "Added"
    summary = summary.rstrip(".")
    if summary:
        summary = summary[0].upper() + summary[1:]
    return heading, f"{summary}."


def insert_changelog(version: str, title: str, body: str, release_date: str) -> None:
    path = Path("CHANGELOG.md")
    text = path.read_text(encoding="utf-8")
    marker = f"## [{version}]"
    if marker in text:
        return

    heading, entry = changelog_section(title, body)
    block = f"## [{version}] - {release_date}\n\n### {heading}\n\n- {entry}\n\n"
    first_release = re.search(r"^## \[\d+\.\d+\.\d+\]", text, re.MULTILINE)
    if first_release:
        text = text[: first_release.start()] + block + text[first_release.start() :]
    else:
        text = text.rstrip() + "\n\n" + block
    path.write_text(text, encoding="utf-8")


def command_write(args: argparse.Namespace) -> None:
    expected = next_version(args.base_version, args.title, args.body)
    Path("VERSION").write_text(expected + "\n", encoding="utf-8")
    if expected != args.base_version:
        insert_changelog(expected, args.title, args.body, args.date)
    print(expected)


def command_check(args: argparse.Namespace) -> None:
    expected = next_version(args.base_version, args.title, args.body)
    actual = Path("VERSION").read_text(encoding="utf-8").strip()
    parse_version(actual)

    if actual != expected:
        raise SystemExit(
            f"VERSION is {actual}, but PR title requires {expected} from base {args.base_version}. "
            "Run scripts/release-version.py write with the same title and base version."
        )

    if expected != args.base_version:
        changelog = Path("CHANGELOG.md").read_text(encoding="utf-8")
        if f"## [{expected}]" not in changelog:
            raise SystemExit(f"CHANGELOG.md does not contain a {expected} release section.")

    print(json.dumps({"base": args.base_version, "expected": expected, "actual": actual}))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser()
    parser.add_argument("command", choices=("write", "check"))
    parser.add_argument("--base-version", required=True)
    parser.add_argument("--title", required=True)
    parser.add_argument("--body", default="")
    parser.add_argument("--date", default=date.today().isoformat())
    return parser


if __name__ == "__main__":
    arguments = build_parser().parse_args()
    parse_version(arguments.base_version)
    if arguments.command == "write":
        command_write(arguments)
    else:
        command_check(arguments)
