FROM l.gcr.io/google/bazel:latest as builder

RUN apt-get update && apt-get install ca-certificates tzdata && update-ca-certificates

ENV USER=ito
ENV UID=10001

RUN adduser \
    --disabled-password \
    --gecos "" \
#    --home "/nonexistent" \
#    --shell "/sbin/nologin" \
#    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /src/ito/api-backend/
COPY . .

#RUN bazel build main && cp $(bazel info bazel-genfiles)/main_/main /bin/backend

#FROM scratch

#COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
#COPY --from=builder /etc/passwd /etc/passwd
#COPY --from=builder /etc/group /etc/group
#COPY --from=builder /bin/backend /go/bin/backend


USER ${USER}:${USER}

ENTRYPOINT ["bazel", "run", "main", "--", "--port", "5628"]
