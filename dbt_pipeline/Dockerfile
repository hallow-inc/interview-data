FROM python:3.13-slim-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends && \
    rm -rd /var/lib/apt/lists/* && \
    mkdir -p $HOME/results

COPY --from=ghcr.io/astral-sh/uv:0.8.10 /uv /uvx /bin/

ADD . /app

WORKDIR /app

RUN rm -rf .venv
ENV UV_LINK_MODE=copy

RUN uv sync --locked

CMD ["tail", "-f", "/dev/null"]
