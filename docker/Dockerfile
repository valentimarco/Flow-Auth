FROM golang:1.22 as builder 

ARG UID
ARG GID
RUN groupadd -g $GID appgroup && \
    useradd -u $UID -g $GID -m appuser
USER appuser

WORKDIR /home/appuser/app

COPY cmd ./cmd
COPY go.mod go.sum ./
RUN go mod download && go mod verify

RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./server ./cmd

FROM scratch
COPY --from=builder /home/appuser/app/server /server
EXPOSE 3000
ENTRYPOINT ["/server"]