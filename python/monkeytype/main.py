#!/usr/bin/env python3
"""
Auto Type Hint Generator for Python
Uses MonkeyType + pytest to trace runtime types and apply annotations
"""

import subprocess
import sys
import json
from pathlib import Path
from dataclasses import dataclass
from typing import Optional


@dataclass
class TypeHintResult:
    module: str
    success: bool
    types_added: int
    output: str
    suggestions: list[str]


def run_command(cmd: list[str], cwd: Optional[Path] = None, timeout: int = 300) -> tuple[bool, str]:
    """Run a command and return success status and output."""
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        output = result.stdout + result.stderr
        return result.returncode == 0, output
    except subprocess.TimeoutExpired:
        return False, "Command timed out"
    except FileNotFoundError as e:
        return False, f"Tool not found: {e}"


def check_monkeytype_installed() -> bool:
    """Check if MonkeyType is installed."""
    success, _ = run_command(["monkeytype", "--version"])
    return success


def install_dependencies() -> bool:
    """Install required dependencies."""
    print("Installing dependencies...")
    success, output = run_command([
        "pip", "install", 
        "monkeytype", 
        "pytest", 
        "pytest-monkeytype",
        "libcst"
    ])
    
    if success:
        print("✅ Dependencies installed")
    else:
        print(f"❌ Failed to install: {output}")
    
    return success


def run_tests_with_tracing(test_path: Path, src_path: Path) -> TypeHintResult:
    """Run pytest with MonkeyType tracing enabled."""
    print(f"\nRunning tests with type tracing: {test_path}")
    
    # Method 1: Using pytest-monkeytype plugin
    success, output = run_command([
        "pytest",
        "--monkeytype-output=monkeytype.db",
        "--monkeytype",
        str(test_path),
    ], cwd=src_path.parent)
    
    if not success:
        # Fallback: Run with monkeytype run
        print("  Falling back to monkeytype run...")
        success, output = run_command([
            "monkeytype",
            "run",
            "-m", "pytest",
            str(test_path),
        ], cwd=src_path.parent)
    
    return TypeHintResult(
        module=str(test_path),
        success=success,
        types_added=0,  # Will be updated after apply
        output=output,
        suggestions=[
            "Ensure test coverage is comprehensive",
            "Avoid heavy mocking which can obscure real types",
            "Run tests multiple times with different inputs for better coverage"
        ]
    )


def apply_type_hints(module_name: str, src_path: Path) -> TypeHintResult:
    """Apply collected type hints to source code."""
    print(f"\nApplying type hints to: {module_name}")
    
    # First, check what types were collected
    success, list_output = run_command([
        "monkeytype",
        "list-modules",
    ], cwd=src_path.parent)
    
    if not success or not list_output.strip():
        return TypeHintResult(
            module=module_name,
            success=False,
            types_added=0,
            output="No type information collected. Run tests first.",
            suggestions=[
                "Run tests with MonkeyType tracing enabled",
                "Ensure the module is imported during test execution",
                "Check that tests actually call the functions"
            ]
        )
    
    # Apply types to the module
    success, apply_output = run_command([
        "monkeytype",
        "apply",
        module_name,
    ], cwd=src_path.parent)
    
    # Count types added (rough estimate from output)
    types_added = apply_output.count("->") + apply_output.count(": ")
    
    suggestions = []
    if success:
        suggestions.extend([
            "Review applied types for accuracy",
            "Run mypy to validate type correctness",
            "Consider adding Union types for functions with multiple return types"
        ])
    else:
        suggestions.extend([
            "Check for conflicting type information",
            "Ensure all test paths have consistent types",
            "Manually add types for complex cases"
        ])
    
    return TypeHintResult(
        module=module_name,
        success=success,
        types_added=types_added,
        output=apply_output,
        suggestions=suggestions
    )


def generate_stub(module_name: str, src_path: Path) -> TypeHintResult:
    """Generate .pyi stub file instead of modifying source."""
    print(f"\nGenerating stub for: {module_name}")
    
    success, output = run_command([
        "monkeytype",
        "stub",
        module_name,
    ], cwd=src_path.parent)
    
    stub_file = src_path.parent / f"{module_name.replace('.', '/')}.pyi"
    
    if success and output.strip():
        # Write stub file
        stub_file.parent.mkdir(parents=True, exist_ok=True)
        stub_file.write_text(output)
        output = f"Stub written to: {stub_file}\n\n{output}"
    
    return TypeHintResult(
        module=module_name,
        success=success,
        types_added=1 if success else 0,
        output=output,
        suggestions=[
            "Review stub file for accuracy",
            "Import stub in your code: from . import module",
            "Use stubs for gradual typing adoption"
        ]
    )


def run_autotyping(src_path: Path) -> TypeHintResult:
    """Use autotyping (ruff-based) to add obvious type hints."""
    print("\nRunning autotyping (static inference)...")
    
    success, output = run_command([
        "ruff",
        "check",
        "--select=UP006,UP007",  # Use list/dict type hints
        "--fix",
        str(src_path),
    ])
    
    # Also run pyupgrade for type-related upgrades
    success2, output2 = run_command([
        "ruff",
        "check",
        "--select=UP",
        "--fix",
        str(src_path),
    ])
    
    return TypeHintResult(
        module=str(src_path),
        success=success or success2,
        types_added=output.count("Fixed"),
        output=output + "\n" + output2,
        suggestions=[
            "Run MonkeyType for runtime-based type inference",
            "Use pyright --writeinfo for additional static inference",
            "Review Union types for Optional conversions"
        ]
    )


def validate_with_mypy(src_path: Path) -> tuple[bool, str]:
    """Validate applied types with mypy."""
    print("\nValidating with mypy...")
    
    success, output = run_command([
        "mypy",
        "--strict",
        "--show-error-codes",
        str(src_path),
    ])
    
    return success, output


def scan_and_annotate(src_path: Path, test_path: Optional[Path] = None, use_stubs: bool = False) -> None:
    """Main workflow: scan, trace, and annotate."""
    
    # Check/install dependencies
    if not check_monkeytype_installed():
        print("MonkeyType not found. Installing...")
        if not install_dependencies():
            print("Failed to install dependencies. Exiting.")
            sys.exit(1)
    
    print(f"\nScanning: {src_path}")
    print("=" * 60)
    
    results = []
    
    # Step 1: Static inference with autotyping
    print("\n[Step 1/4] Running static type inference...")
    result = run_autotyping(src_path)
    results.append(("Static (autotyping)", result))
    
    # Step 2: Run tests with tracing
    if test_path:
        print("\n[Step 2/4] Running tests with type tracing...")
        result = run_tests_with_tracing(test_path, src_path)
        results.append(("Test tracing", result))
    else:
        print("\n[Step 2/4] Skipping test tracing (no test path provided)")
        print("  Provide test path with --tests to enable runtime tracing")
    
    # Step 3: Apply or generate stubs
    print("\n[Step 3/4] Applying type hints...")
    
    # Find all Python modules
    if src_path.is_file():
        modules = [src_path]
    else:
        modules = list(src_path.rglob("*.py"))
    
    for module in modules[:10]:  # Limit to first 10 for demo
        if "test" in str(module) or "__pycache__" in str(module):
            continue
        
        module_name = str(module.relative_to(src_path.parent)).replace("/", ".").replace(".py", "")
        
        if use_stubs:
            result = generate_stub(module_name, src_path)
        else:
            result = apply_type_hints(module_name, src_path)
        
        results.append((f"Apply: {module_name}", result))
    
    # Step 4: Validate
    print("\n[Step 4/4] Validating with mypy...")
    mypy_success, mypy_output = validate_with_mypy(src_path)
    results.append(("mypy validation", TypeHintResult(
        module="all",
        success=mypy_success,
        types_added=0,
        output=mypy_output,
        suggestions=[]
    )))
    
    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    
    for name, result in results:
        status = "✅" if result.success else "⚠️"
        print(f"{status} {name}: {result.types_added} types added")
    
    # Generate report
    report = []
    report.append("# Type Hint Annotation Report")
    report.append(f"Generated: {subprocess.run(['date'], capture_output=True, text=True).stdout.strip()}")
    report.append("")
    
    for name, result in results:
        report.append(f"## {name}")
        report.append("")
        if result.output.strip():
            report.append("```")
            report.append(result.output[:2000])
            report.append("```")
        if result.suggestions:
            report.append("**Suggestions:**")
            for s in result.suggestions:
                report.append(f"- {s}")
        report.append("")
    
    report_file = src_path.parent / "type_hints_report.md"
    report_file.write_text("\n".join(report))
    print(f"\nReport saved to: {report_file}")


def main():
    if len(sys.argv) < 2:
        print("Usage: python auto_type_hints.py <src_path> [options]")
        print("  --tests <path>: Path to test directory for runtime tracing")
        print("  --stubs: Generate .pyi stub files instead of modifying source")
        print("  --install-only: Just install dependencies")
        sys.exit(1)
    
    src_path = Path(sys.argv[1]).resolve()
    if not src_path.exists():
        print(f"Error: Path does not exist: {src_path}")
        sys.exit(1)
    
    test_path = None
    use_stubs = False
    
    i = 2
    while i < len(sys.argv):
        if sys.argv[i] == "--tests" and i + 1 < len(sys.argv):
            test_path = Path(sys.argv[i + 1]).resolve()
            i += 2
        elif sys.argv[i] == "--stubs":
            use_stubs = True
            i += 1
        elif sys.argv[i] == "--install-only":
            install_dependencies()
            sys.exit(0)
        else:
            i += 1
    
    scan_and_annotate(src_path, test_path, use_stubs)


if __name__ == "__main__":
    main()
