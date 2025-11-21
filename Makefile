PYTHON ?= python

.PHONY: install-dev test cov cov-xml cov-html

install-dev:
	$(PYTHON) -m pip install --upgrade pip
	pip install -e .[dev]

# Run tests (quiet) with coverage summary in terminal
test:
	$(PYTHON) -m pytest

# Terminal coverage with missing lines
cov:
	$(PYTHON) -m pytest --cov-report=term-missing

# Generate coverage.xml for CI tools
cov-xml:
	$(PYTHON) -m pytest --cov-report=xml

# Generate HTML coverage report (open htmlcov/index.html)
cov-html:
	$(PYTHON) -m pytest --cov-report=html
