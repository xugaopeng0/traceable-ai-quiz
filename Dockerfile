# check=skip=InvalidDefaultArgInFrom
# Base images are required Compose build arguments sourced from .env.
ARG GO_IMAGE
ARG NODE_IMAGE
ARG NGINX_IMAGE
ARG ALPINE_IMAGE

FROM ${GO_IMAGE} AS go-build
WORKDIR /src
COPY backend/ ./
ENV CGO_ENABLED=0
RUN go build -trimpath -o /out/api ./cmd/api && go build -trimpath -o /out/web-log ./cmd/web-log

FROM go-build AS backend-test
RUN go test ./... && go vet ./...

FROM ${NODE_IMAGE} AS frontend-build
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

FROM ${ALPINE_IMAGE} AS api
COPY --from=go-build /out/api /app/api
USER 65532:65532
ENTRYPOINT ["/app/api"]

FROM ${NGINX_IMAGE} AS web
STOPSIGNAL SIGTERM
COPY --from=go-build /out/web-log /usr/local/bin/web-log
COPY --from=frontend-build /src/dist/ /usr/share/nginx/html/
USER 101:101
ENTRYPOINT ["/usr/local/bin/web-log"]
