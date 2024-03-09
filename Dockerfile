ARG GOLANG_VERSION="1.19.1"

FROM golang:$GOLANG_VERSION-alpine as builder
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-s' -o ./app

FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tencent.com/g' /etc/apk/repositories
ENV TZ=Asia/Shanghai
RUN apk --no-cache add tzdata && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime && echo "${TZ}" > /etc/timezone
COPY --from=builder /app/app /
ENTRYPOINT ["/app"]
