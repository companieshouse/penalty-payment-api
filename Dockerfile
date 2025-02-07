FROM 169942020521.dkr.ecr.eu-west-2.amazonaws.com/base/golang:1.19-bullseye-builder AS builder

RUN /bin/go_build

FROM 169942020521.dkr.ecr.eu-west-2.amazonaws.com/base/golang:debian11-runtime

COPY --from=builder /build/out/app ./

COPY assets ./assets

CMD ["-bind-addr=:4086"]

EXPOSE 4086
