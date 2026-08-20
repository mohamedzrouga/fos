#!/usr/bin/env python3
"""
Code Quality Scanner for Python
Scans for: linting (Ruff), type checking (mypy), performance issues, security (Bandit)
"""

import subprocess
import sys
import json
from pathlib import Path
from dataclasses import dataclass
from typing import Optional


@dataclass
class ScanResult:
    tool: str
    success: bool
    output: str
    suggestions: list[str]


def run_command(cmd: list[str], cwd: Optional[Path] = None) -> tuple[bool, str]:
    """Run a command and return success status and output."""
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=300,
        )
        output = result.stdout + result.stderr
        return result.returncode == 0, output
    except subprocess.TimeoutExpired:
        return False, "Command timed out"
    except FileNotFoundError as e:
        return False, f"Tool not found: {e}"


def check_ruff(path: Path) -> ScanResult:
    """Run Ruff for linting and code quality."""
    # Check with all rules enabled
    success, output = run_command([
        "ruff", "check", "--output-format=full", "--select=ALL", str(path)
    ])
    
    suggestions = []
    if not success:
        suggestions.extend([
            "Run 'ruff check --fix' to auto-fix issues",
            "Add # noqa: <CODE> for intentional violations",
            "Configure rules in pyproject.toml under [tool.ruff]"
        ])
    
    return ScanResult("Ruff (linting)", success, output, suggestions)


def check_ruff_performance(path: Path) -> ScanResult:
    """Run Ruff performance-specific checks."""
    success, output = run_command([
        "ruff", "check", 
        "--select=PERF,UP,C4,SIM",  # Performance, Upgrade, Comprehensions, Simplify
        str(path)
    ])
    
    suggestions = []
    if "PERF" in output or "UP" in output:
        suggestions.extend([
            "Consider using list comprehensions instead of map/filter",
            "Use f-strings instead of .format() or %",
            "Avoid unnecessary list() calls on already iterable objects",
            "Use 'in' for membership tests instead of manual loops"
        ])
    
    return ScanResult("Ruff (performance)", success, output, suggestions)


def check_mypy(path: Path) -> ScanResult:
    """Run mypy for type checking."""
    success, output = run_command([
        "mypy", 
        "--strict",
        "--show-error-codes",
        "--pretty",
        str(path)
    ])
    
    suggestions = []
    if not success:
        suggestions.extend([
            "Add type hints to function parameters and return values",
            "Use Optional[X] or X | None for nullable types",
            "Consider using TypedDict for dictionary structures",
            "Run 'mypy --ignore-missing-imports' for third-party libs without stubs"
        ])
    
    return ScanResult("mypy (types)", success, output, suggestions)


def check_bandit(path: Path) -> ScanResult:
    """Run Bandit for security issues."""
    success, output = run_command([
        "bandit", 
        "-r", str(path),
        "-f", "custom",
        "-o", "stdout"
    ])
    
    suggestions = []
    if not success:
        suggestions.extend([
            "Avoid using eval(), exec(), or compile() with user input",
            "Use parameterized queries instead of string formatting for SQL",
            "Don't hardcode passwords or secrets in code",
            "Use hashlib instead of md5/sha1 for security-critical hashing"
        ])
    
    return ScanResult("Bandit (security)", success, output, suggestions)


def check_pylint_deep(path: Path) -> ScanResult:
    """Run Pylint for deep semantic analysis (optional, slower)."""
    success, output = run_command([
        "pylint",
        "--output-format=text",
        "--reports=no",
        "--score=no",
        "--disable=C,R",  # Skip convention and refactoring for brevity
        str(path)
    ])
    
    suggestions = []
    if "W" in output or "E" in output:
        suggestions.extend([
            "Check for unused imports and variables",
            "Review duplicate code patterns",
            "Ensure proper exception handling"
        ])
    
    return ScanResult("Pylint (deep)", success, output, suggestions)


def generate_report(results: list[ScanResult], output_file: Path) -> None:
    """Generate a summary report."""
    report_lines = [
        "# Code Quality Scan Report",
        f"Generated: {subprocess.run(['date'], capture_output=True, text=True).stdout.strip()}",
        "",
        "## Summary",
        "",
    ]
    
    for result in results:
        status = "✅ PASS" if result.success else "❌ ISSUES FOUND"
        report_lines.append(f"### {result.tool}: {status}")
        report_lines.append("")
        
        if result.output.strip():
            # Truncate long outputs
            output_preview = result.output[:2000]
            if len(result.output) > 2000:
                output_preview += "\n... (truncated)"
            report_lines.append("```")
            report_lines.append(output_preview)
            report_lines.append("```")
            report_lines.append("")
        
        if result.suggestions:
            report_lines.append("**Suggestions:**")
            for suggestion in result.suggestions:
                report_lines.append(f"- {suggestion}")
            report_lines.append("")
    
    output_file.write_text("\n".join(report_lines))
    print(f"Report saved to: {output_file}")


def main():
    if len(sys.argv) < 2:
        print("Usage: python scan_code.py <path> [--deep]")
        print("  --deep: Include Pylint for deeper analysis (slower)")
        sys.exit(1)
    
    target_path = Path(sys.argv[1]).resolve()
    if not target_path.exists():
        print(f"Error: Path does not exist: {target_path}")
        sys.exit(1)
    
    use_deep = "--deep" in sys.argv
    
    print(f"Scanning: {target_path}")
    print("=" * 60)
    
    results: list[ScanResult] = []
    
    # Run checks
    checks = [
        ("Ruff (linting)", lambda: check_ruff(target_path)),
        ("Ruff (performance)", lambda: check_ruff_performance(target_path)),
        ("mypy (types)", lambda: check_mypy(target_path)),
        ("Bandit (security)", lambda: check_bandit(target_path)),
    ]
    
    if use_deep:
        checks.append(("Pylint (deep)", lambda: check_pylint_deep(target_path)))
    
    for name, check_fn in checks:
        print(f"\nRunning {name}...")
        try:
            result = check_fn()
            results.append(result)
            status = "✅ PASS" if result.success else "⚠️  ISSUES"
            print(f"  {status}")
        except Exception as e:
            print(f"  ❌ ERROR: {e}")
            results.append(ScanResult(name, False, str(e), []))
    
    # Generate report
    report_file = target_path / "code_quality_report.md" if target_path.is_dir() else target_path.parent / "code_quality_report.md"
    generate_report(results, report_file)
    
    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    passed = sum(1 for r in results if r.success)
    print(f"Passed: {passed}/{len(results)}")
    
    if passed < len(results):
        print("\nReview the report for detailed suggestions.")
        sys.exit(1)


if __name__ == "__main__":
    main()
