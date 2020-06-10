FROM golang:1.14.3 AS prod
WORKDIR /app
COPY main.go go.mod .envstage sounds/ ./
RUN apt-get update && \
    apt-get install -y ffmpeg
ENV GO_ENV=stage
RUN go install .
ENTRYPOINT ["kaboom"]