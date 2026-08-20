# Setup and Usage
### Python

```bash
# Install dependencies
pip install monkeytype pytest pytest-monkeytype libcst mypy ruff

# Or using uv
uv pip install monkeytype pytest pytest-monkeytype libcst mypy ruff

# Run the auto-annotator
python auto_type_hints.py /path/to/your/code

# With test tracing (recommended)
python auto_type_hints.py /path/to/your/code --tests /path/to/tests

# Generate stub files instead of modifying source
python auto_type_hints.py /path/to/your/code --stubs

# Just install dependencies
python auto_type_hints.py /path/to/your/code --install-only
```

**Example workflow:**

```bash
# 1. Run your tests to collect type information
monkeytype run pytest tests/

# 2. Apply types to a specific module
monkeytype apply mypackage.mymodule

# 3. Or generate a stub file
monkeytype stub mypackage.mymodule > mypackage/mymodule.pyi

# 4. Validate with mypy
mypy --strict mypackage/
```


## pytest-monkeytype Plugin Configuration

Create `pytest.ini` or add to `pyproject.toml`:

```toml
# pyproject.toml
[tool.pytest.ini_options]
addopts = """
  --monkeytype
  --monkeytype-output=monkeytype.db
  --monkeytype-apply
"""

[tool.monkeytype]
# Optional configuration
disable = []
limit = 0
```
