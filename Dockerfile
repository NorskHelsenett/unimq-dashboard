ARG GCR_MIRROR=gcr.io/

FROM ${GCR_MIRROR}distroless/nodejs26-debian13

LABEL org.opencontainers.image.source "https://github.com/norskhelsenett/unimq-dashboard"

WORKDIR /app

COPY /web/static/dist/* ./web/dist/

CMD ["/app/web/dist/index.js"]
