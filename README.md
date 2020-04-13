# ito - Upload API

Public API for supplying and verifying pseudonyms of users confirmed infected

![Tests](https://github.com/ito-org/upload-api/workflows/Build/badge.svg)

## Requirements

- Python 3.8+
- MongoDB database

<details>
  <summary>Quick start Python 3.8 (Debian based Linux)</summary>

```bash
sudo apt install python3.8 python3.8-pip
sudo update-alternatives --config python3
```

Then select the correct Python version.

</details>

## Installation

Install and initialize [Poetry](https://python-poetry.org/docs). Run

```bash
poetry install
```

<details>
  <summary>Quick start Poetry (UNIX)</summary>

```bash
curl -sSL https://raw.githubusercontent.com/python-poetry/poetry/master/get-poetry.py | python
source ~/.poetry/env
```

</details>

## Configuration

Copy the `config.py.example` to `config.py` and adjust the database connection URI.

## Development

Run the local Flask development server using

```bash
export POETRY_ENV="development"
poetry run flask run
```

Then send a POST request to http://localhost:5000/v0/cases/report for example.
