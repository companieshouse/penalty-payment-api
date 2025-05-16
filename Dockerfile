FROM golang:1.24.2-bullseye AS builder

ENV GOPRIVATE="github.com/companieshouse"

ARG SSH_PRIVATE_KEY
ARG SSH_PRIVATE_KEY_PASSPHRASE

COPY ./bin/go_build /bin/

RUN chmod +x /bin/go_build && \
    git config --global url."git@github.com:".insteadOf https://github.com/

WORKDIR /build

COPY . /build/

RUN /bin/go_build

FROM gcr.io/distroless/base-debian11:latest@sha256:ac69aa622ea5dcbca0803ca877d47d069f51bd4282d5c96977e0390d7d256455 AS runner

WORKDIR /app

COPY --from=builder /build/out/app ./

COPY assets ./assets

CMD ["-bind-addr=:4086"]

EXPOSE 4086

USER nonroot:nonroot

ENTRYPOINT ["/app/app"]
