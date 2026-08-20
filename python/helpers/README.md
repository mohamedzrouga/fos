## Setup Instructions

### Python (requires Python 3.10+)

```bash
# Install dependencies
pip install ruff mypy bandit pylint

# Or using uv (faster)
uv pip install ruff mypy bandit pylint

# Run the scanner
python scan_code.py /path/to/your/code

# For deeper analysis (includes Pylint)
python scan_code.py /path/to/your/code --deep
```

**Optional: Add to `pyproject.toml` for configuration:**

```toml
[tool.ruff]
line-length = 88
target-version = "py310"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "N", "UP", "B", "C4", "SIM", "PERF"]

[tool.mypy]
python_version = "3.10"
strict = true
warn_return_any = true
disallow_untyped_defs = true
```
