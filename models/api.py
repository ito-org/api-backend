import json
from flask import Response
from typing import Optional


class APIError:
    code: int
    message: str

    def __init__(self, code: Optional[int] = None, message: Optional[str] = None):
        if code is None:
            self.code = 500
        else:
            self.code = code

        if message is None:
            self.message = "Some unexpected condition occurred"
        else:
            self.message = message

    def as_response(self) -> Response:
        return Response(
            json.dumps(self.__dict__), status=self.code, mimetype="application/json"
        )
