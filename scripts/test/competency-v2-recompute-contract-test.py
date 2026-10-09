#!/usr/bin/env python3
"""Regression contract for the staging v2 recompute verifier."""

import ast
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / "scripts" / "tools" / "staging-competency-v2-recompute.py"


def is_generate_report_branch(node: ast.If) -> bool:
    test = node.test
    return (
        isinstance(test, ast.Attribute)
        and isinstance(test.value, ast.Name)
        and test.value.id == "args"
        and test.attr == "generate_report"
    )


def main() -> None:
    tree = ast.parse(SCRIPT.read_text(encoding="utf-8"), filename=str(SCRIPT))
    branches = [node for node in ast.walk(tree) if isinstance(node, ast.If) and is_generate_report_branch(node)]
    assert len(branches) == 1, "expected one --generate-report branch"
    assert not branches[0].orelse, (
        "FB-189: default recompute verification must not assume that v2 report generation remains disabled"
    )
    print("COMPETENCY_V2_RECOMPUTE_CONTRACT_TEST_PASS")


if __name__ == "__main__":
    main()
