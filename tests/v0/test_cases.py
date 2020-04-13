import json
from datetime import datetime
from flask import url_for, Response
from flask.testing import FlaskClient


def test_report_post(client: FlaskClient):
    default_reportsig = "teststr"

    data = {
        "reportsig": default_reportsig,
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
    }
    res: Response = client.post(
        url_for(".report"), data=json.dumps(data), content_type="application/json"
    )
    assert res.status_code == 200


def test_report_post_duplicate(client: FlaskClient):
    duplicate_reportsig = "duplicate"

    data1 = {
        "reportsig": duplicate_reportsig,
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
    }
    res1: Response = client.post(
        url_for(".report"), data=json.dumps(data1), content_type="application/json"
    )
    assert res1.status_code == 200

    data2 = {
        "reportsig": duplicate_reportsig,
        "timestamp": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
    }
    res2: Response = client.post(
        url_for(".report"), data=json.dumps(data2), content_type="application/json"
    )
    assert res2.status_code == 200
