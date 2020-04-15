from flask import Flask
from pymongo import MongoClient, DESCENDING  # type: ignore
from uuid import uuid4, UUID
from datetime import datetime
import time
from random import randrange


class DBConnection:
    def __init__(self, hostUri):
        if hostUri is None:
            hostUri = "mongodb://localhost:27017"
        client = MongoClient(hostUri)
        self.db = client.ito

        # Create index for reportsigs collection to make sure that reportsig
        # is unique.
        self.db.reportsigs.create_index([("reportsig", DESCENDING)], unique=True)

    def insert_reportsig(self, reportsig, timestamp) -> None:
        self.db.reportsigs.insert_one({"reportsig": reportsig, "timestamp": timestamp})

    def get_reportsigs(self):
        return list(self.db.reportsigs.find({}, {"_id": False}))
