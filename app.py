from flask import Flask
import os
import json
from pymongo.errors import DuplicateKeyError  # type: ignore
from flask import Flask, Response, request
from db.db import DBConnection
from models.api import APIError

app = Flask(__name__)
dbConn = DBConnection(os.environ.get("MONGO_URI"))


@app.route("/report", methods=["GET", "POST"])
def report() -> Response:
    if request.method == "GET":
        return Response(
            json.dumps(dbConn.get_reportsigs()), 200, mimetype="application/json"
        )
    else:
        data = request.get_json()
        if "reportsig" not in data or "timestamp" not in data:
            return APIError(400, "Missing values in request").as_response()
        reportsig = data["reportsig"]
        timestamp = data["timestamp"]

        try:
            dbConn.insert_reportsig(reportsig, timestamp)
        except:
            pass
        return Response(None, 200)
