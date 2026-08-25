# dream-narrative-analysis-workbench-117

基于 Go 实现的梦境叙事与睡眠线索分析工作台 Web 项目，一款后端服务，这是一个面向个人用户的 Go 全栈梦境观察工具，它把醒来后的片段记录、梦中元素、重复主题、醒来情绪和睡眠时长放到同一条可回看的时间线上，帮助用户发现自己的叙事规律，系统只提供可解释的观察线索，不替代医学诊断，也不会把主题命中当作固定心理含义。

项目源代码、依赖描述和评测专用 Docker 文件共同构成自包含任务；不依赖本机预编译二进制。

## 标准构建、运行和测试命令

```bash
go build ./...
go run ./cmd/server
go test ./...
```
## 评测容器

评测专用 Dockerfile 为 `benzhi.Dockerfile`，构建脚本为 `build_benzhi_docker.sh`。

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh my-go-task linux/arm64
./build_benzhi_docker.sh my-go-task linux/amd64
docker run -it my-go-task:latest
```
