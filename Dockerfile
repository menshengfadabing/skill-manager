# syntax=docker/dockerfile:1

FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/skill ./cmd/skill

FROM alpine:3.20
RUN apk add --no-cache ca-certificates git
COPY --from=build /out/skill /usr/local/bin/skill
COPY skills /usr/local/share/skill-manager/skills
ENV SKILL_MANAGER_BUNDLED=/usr/local/share/skill-manager/skills
WORKDIR /work
ENTRYPOINT ["skill"]
CMD ["help"]
