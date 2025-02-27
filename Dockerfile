FROM golang:1.22-alpine AS builder

WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum* ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# 使用更小的基础镜像
FROM alpine:latest

WORKDIR /app

# 从builder阶段复制编译好的应用
COPY --from=builder /app/app .

# 复制静态文件
COPY --from=builder /app/static ./static

# 设置环境变量
ENV PORT=8080

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./app"] 