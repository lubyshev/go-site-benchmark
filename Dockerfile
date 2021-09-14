FROM golang as builder
MAINTAINER Nick Lubyshev <lubyshev@gmail.com>

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app ./main.go

FROM alpine
RUN mkdir /runtime;mkdir /runtime/etc;mkdir /runtime/certs
WORKDIR /runtime
COPY --from=builder /build/app /runtime/app
COPY --from=builder /build/etc /runtime/etc
COPY --from=builder /build/certs /runtime/certs

ENTRYPOINT [ "/runtime/app" ]
