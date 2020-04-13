FROM python:3.8

ARG INSTALL_PATH="/ito"

ENV PIP_DISABLE_PIP_VERSION_CHECK=on
ENV FLASK_ENV "development"
ENV MONGO_URI "mongodb://localhost:27017"

WORKDIR ${INSTALL_PATH}

RUN pip install poetry

COPY . ${INSTALL_PATH}/

RUN mv config.py.example config.py && \ 
    poetry config virtualenvs.create false && \
    poetry install

CMD poetry run flask run --host 0.0.0.0 --port=5001