from flask import Flask
import pytest  # type: ignore
from app import app as theapp


@pytest.fixture  # type: ignore
def app() -> Flask:
    return theapp
