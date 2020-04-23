# ito - Backend API

API for supplying and verifying TCNs of users confirmed infected

<!--![Tests](https://github.com/ito-org/api-backend/workflows/Build/badge.svg)-->
[![Tests](https://img.shields.io/badge/tests-work%20in%20progress%20-black)](LICENSE)
[![License](https://img.shields.io/badge/license-BSD--3--Clause--Clear-brightgreen)](LICENSE)

## Prerequisites

- Go
- PostgreSQL

## Run it

Run the backend directly by spinning up a [Postgres Docker](https://hub.docker.com/_/postgres/) container and running `go run github.com/ito-org/api-backend`. Alternative, you can spin up the backend in combination with the database via docker-compose. Run `docker-compose build && docker-compose up -d`.

**IMPORTANT**: Keep in mind that you need to set the environment variables as shown below.

## Environment variables

You can supply database credentials through environment variables. The following are available:

* `POSTGRES_DB`
* `POSTGRES_USER`
* `POSTGRES_PASSWORD`

You can either set them directly when running the application or set them through an `.env` file in the project root. For docker-compose, the `.env` file is required.