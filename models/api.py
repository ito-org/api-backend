import json
from flask import Response


class APIError:
    def __init__(self, code, message):
        if code is None:
            self.code = 500
        else:
            self.code = code

        if message is None:
            self.message = "Some unexpected condition occurred"
        else:
            self.message = message

    def as_response(self):
        return Response(
            json.dumps(self.__dict__), status=self.code, mimetype="application/json"
        )
