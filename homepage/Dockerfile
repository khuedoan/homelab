FROM rust:1.44.1-slim AS base

ENV USER=root

WORKDIR /usr/local/src/homepage/

RUN cargo init
COPY ./Cargo.toml ./Cargo.toml
RUN cargo fetch

COPY ./src/ ./src/

CMD [ "cargo", "test", "--offline" ]

FROM base AS builder

RUN cargo build --release --offline

FROM rust:1.44.1-slim

COPY --from=builder /usr/local/src/homepage/target/release/homepage /usr/local/bin/homepage

EXPOSE 8000

ENTRYPOINT [ "/usr/local/bin/homepage" ]
