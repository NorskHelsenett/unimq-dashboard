ARG GCR_MIRROR=gcr.io/

FROM ${GCR_MIRROR}distroless/nodejs26-debian13
# FROM ubuntu:noble

# LABEL org.opencontainers.image.source https://github.com/norskhelsenett/unimq-dashboard

WORKDIR /app

COPY /web/* ./web/

CMD ["/app/web/templates/index.html"]
