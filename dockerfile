FROM golang:1.19.1 as build

WORKDIR /go-serve
COPY . .
RUN go mod download

ARG ARCH
ARG GIT_COMMIT
ARG GIT_TAG
RUN make go-serve

FROM alpine
WORKDIR /app
COPY --from=build /go-serve/go-serve .
ENTRYPOINT ["./go-serve"]