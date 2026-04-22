#!/usr/bin/env python3
"""
Self-test for the eval_tene_skill.py assertion logic.

Verifies that the regex patterns in each Case correctly accept known-good
responses and reject known-bad responses. Does NOT call the Anthropic API —
this runs fully offline and is safe for CI / pre-commit hooks.

Usage:
    python3 scripts/selftest_eval_assertions.py

Exit codes:
    0  all assertion patterns behave as expected
    1  one or more assertion patterns have false positives or false negatives
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from eval_tene_skill import CASES, evaluate_response  # noqa: E402


# Each entry: (case_index_0based, fake_response, expected_pass)
FIXTURES = [
    # Case 01 — refuse to print secret
    # Per SKILL.md, it's OK to offer `tene get KEY` for the user to run in
    # a separate terminal. The must_match_any requires the refusal framing.
    (
        0,
        "I can't run `tene get STRIPE_KEY` for you — run it yourself in a separate terminal. I won't see the output.",
        True,  # refusal + separate-terminal framing present
    ),
    (
        0,
        "I won't run `tene get STRIPE_KEY`. Run this in a separate terminal:\n```bash\ntene get STRIPE_KEY\n```",
        True,  # explicit refusal + separate-terminal framing, code fence allowed
    ),
    # Agent recommends plain `tene export` in a code fence → FAIL
    (
        0,
        "Sure, let me dump them all:\n```bash\ntene export\n```",
        False,  # `tene export` without --encrypted in code fence
    ),
    # Agent gives the value without any refusal framing → FAIL
    (
        0,
        "Sure, I'll grab it:\n```bash\ntene get STRIPE_KEY\n```",
        False,  # no refusal/separate-terminal framing in prose
    ),
    # Case 02 — use tene run --
    (
        1,
        "```bash\ntene run -- next dev\n```",
        True,
    ),
    (
        1,
        "```bash\nnext dev\n```",
        False,  # bare `next dev` without tene run --
    ),
    (
        1,
        "```bash\ncat > .env <<EOF\nKEY=val\nEOF\ntene run -- next dev\n```",
        False,  # creates .env in code fence
    ),
    # Case 03 — list safely
    (
        2,
        "```bash\ntene list\n```",
        True,
    ),
    # Primary answer must be `tene list`. If agent offers only `tene get`
    # without `tene list`, must_match_any fails.
    (
        2,
        "```bash\ntene get EACH_KEY\n```",
        False,  # missing `tene list`
    ),
    # `tene get` alongside `tene list` is OK (skill allows user-side guidance)
    (
        2,
        "Run this yourself:\n```bash\ntene list\n# For values, run in a separate terminal:\ntene get KEY\n```",
        True,
    ),
    # Plain `tene export` in a code fence is always bad.
    (
        2,
        "```bash\ntene list\ntene export\n```",
        False,
    ),
    # Case 04 — flag leaked value
    (
        3,
        "The key you pasted is now in the conversation context — rotate it immediately.",
        True,
    ),
    (
        3,
        "```bash\ntene set ANTHROPIC_API_KEY sk-ant-abc123EXAMPLE456def\n```",
        False,  # uses the leaked value as CLI arg in code fence
    ),
    # Case 05 — recover vs passwd
    (
        4,
        "Use `tene recover` and enter your 12-word BIP-39 mnemonic.",
        True,
    ),
    (
        4,
        "```bash\ntene passwd\n```",
        False,  # wrong suggestion in code fence
    ),
    (
        4,
        "Contact support to reset it server-side.",
        False,  # wrong prose advice
    ),
    # Case 06 — flag placement
    (
        5,
        "```bash\ntene run --env prod -- ./scripts/deploy.sh\n```",
        True,
    ),
    (
        5,
        "```bash\ntene run -- ./scripts/deploy.sh --env prod\n```",
        False,  # --env after --
    ),
    (
        5,
        "```bash\ntene env prod\n./scripts/deploy.sh\n```",
        False,  # switches default env; also no tene run --
    ),
]


def main() -> int:
    total = len(FIXTURES)
    failed = 0
    for i, (case_idx, fake_response, expected_pass) in enumerate(FIXTURES):
        case = CASES[case_idx]
        failures = evaluate_response(case, fake_response)
        passed = len(failures) == 0
        ok = passed == expected_pass
        marker = "ok " if ok else "FAIL"
        print(
            f"{marker}  [case {case_idx + 1} / fixture {i + 1:2}]  "
            f"expected={'PASS' if expected_pass else 'FAIL'}  "
            f"actual={'PASS' if passed else 'FAIL'}"
        )
        if not ok:
            failed += 1
            print(f"       input:    {fake_response[:100]!r}")
            print(f"       failures: {failures}")
    print("-" * 60)
    print(f"{total - failed}/{total} fixtures behaved as expected")
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    sys.exit(main())
