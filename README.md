# Tailwind Converter

一个用于在 HTML 和 Go 代码之间进行转换的工具，支持 Tailwind CSS 类。

## 功能

- HTML 转 Go：将 HTML 代码转换为 Go 代码
- Go 转 HTML：将 Go 代码转换为 HTML 代码
- 支持 Tailwind CSS 类
- 实时预览

## 本地运行

```bash
# 克隆仓库
git clone https://github.com/yourusername/tailwind-converter.git
cd tailwind-converter

# 运行应用
go run main.go
```

然后在浏览器中访问 http://localhost:8080

## 部署到云平台

### Render

1. 在 Render 上创建一个新的 Web Service
2. 连接到您的 GitHub 仓库
3. 设置以下选项：
   - 环境：Go
   - 构建命令：`go build -o app`
   - 启动命令：`./app`

### Railway

1. 在 Railway 上创建一个新项目
2. 连接到您的 GitHub 仓库
3. Railway 会自动检测 Go 项目并部署

### Heroku

1. 在 Heroku 上创建一个新应用
2. 连接到您的 GitHub 仓库或使用 Heroku CLI 部署
3. 应用会自动使用 Procfile 启动

### Fly.io

1. 安装 flyctl：`curl -L https://fly.io/install.sh | sh`
2. 登录：`flyctl auth login`
3. 启动应用：`flyctl launch`
4. 部署应用：`flyctl deploy`

## 环境变量

- `PORT`：应用监听的端口（默认为 8080）

## 许可证

MIT
