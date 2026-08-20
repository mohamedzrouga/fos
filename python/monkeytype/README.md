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
