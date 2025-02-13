FROM alpine:3.21

COPY maf /

ENTRYPOINT ["/maf"]
