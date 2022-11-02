FROM caddy:2.6-builder-alpine AS builder

RUN xcaddy build \
    --with github.com/42wim/caddy-gitea@v0.0.3

FROM caddy:2.6.2

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
