FROM golang:1.21 AS base

RUN adduser \
        --disabled-password \
        --gecos "" \
        --home "/nonexistent" \
        --shell "/sbin/nologin" \
        --no-create-home \
        --uid 65532 \
        small-user

 COPY . /src
 RUN   go build -C /src -x -v -o /main

FROM gcr.io/distroless/static-debian11
COPY --from=base /main .
USER small-user:small-user
CMD ["./main"]