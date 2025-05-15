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

FROM debian:bullseye

RUN apt-get update && \
    apt-get install -y --allow-downgrades --no-install-recommends\
    ca-certificates=20210119 && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir /app/

WORKDIR /app

COPY --from=builder /build/out/app ./

COPY assets ./assets

ENTRYPOINT ["/app/app"]

CMD ["-bind-addr=:4086"]

EXPOSE 4086
