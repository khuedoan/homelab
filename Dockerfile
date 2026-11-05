FROM nixery.dev/shell/mdbook/mdbook-mermaid AS build

WORKDIR /src

COPY . .

RUN mdbook-mermaid install . && mdbook build

FROM docker.io/library/nginx:1.27-alpine

COPY --from=build /src/book /usr/share/nginx/html
