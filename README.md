# 梦境叙事与睡眠线索分析工作台

这是一个面向个人用户的 Go 全栈梦境观察工具。它把醒来后的片段记录、梦中元素、重复主题、醒来情绪和睡眠时长放到同一条可回看的时间线上，帮助用户发现自己的叙事规律。系统只提供可解释的观察线索，不替代医学诊断，也不会把主题命中当作固定心理含义。

## 问题背景

梦境记忆通常在醒来后快速褪色，事后再回看又很难比较不同日期的情绪、睡眠和叙事元素。本项目提供一个轻量的个人档案：先保存原始叙事，再用证据词匹配主题，用统计窗口比较睡眠与清晰度，最后生成每周可读的回顾信。

## 目标用户与术语

- 目标用户：希望持续记录梦境的个人用户、关注睡眠与情绪关系的自我观察者、整理创作灵感的写作者。
- 梦境记录：一次醒来后的标题、正文、梦境日期、醒来时间、睡眠时长、清晰度和情绪。
- 元素：人物、动物、地点、物品、动作、颜色、自然七类叙事对象。
- 主题线索：由证据词命中的叙事模式，例如“被追赶”“坠落”“迷路”；它是可复核线索，不是解释结论。
- 分析窗口：默认近 30 天，也可以切换到 7 天、90 天或自定义日期范围。

## 核心流程

1. 注册或登录个人工作台，服务使用签名 Cookie 会话。
2. 在“醒来速记”填写标题、叙事、日期、醒来时间、睡眠时长、情绪和清晰度。
3. 系统从正文推荐元素并匹配主题证据词；用户可以点击调整元素或添加自定义元素。
4. 时间线支持按日期和关键词回看，点击卡片可以编辑，服务会重新计算主题并让旧周报进入待刷新状态。
5. 分析页展示主题频次、情绪时间序列、词云、睡眠切片、睡眠/清晰度相关系数和解释性提示。
6. 周报由后台刷新队列和手动刷新接口生成，可导出 Markdown 阅读稿或 JSON 备份。

## 模块说明

- `internal/domain`：领域对象、情绪/元素枚举和输入不变量。
- `internal/store`：带锁的 JSON 持久化，提供用户、梦境、元素、周报和分析任务查询。
- `internal/text`：中文双字/三字词提取、停用词处理、元素推荐和主题证据匹配。
- `internal/analysis`：时间窗口、主题统计、情绪趋势、睡眠相关计算、词云和周报生成。
- `internal/service`：认证、梦境写入、元素计数、分析、周报和导出应用服务。
- `internal/httpapi`：会话保护、输入解码、CORS、路由、API 响应和下载接口。
- `internal/jobs`：刷新任务队列、领取、失败状态和重试调度。
- `web`：原生 HTML、CSS、JavaScript 浏览器工作台，由 Go `embed` 打包进二进制。

## API 摘要

认证：`POST /api/v1/auth/register`、`POST /api/v1/auth/login`、`GET /api/v1/session`、`POST /api/v1/auth/logout`。

梦境：`POST /api/v1/dreams`、`GET /api/v1/dreams`、`GET/PUT/DELETE /api/v1/dreams/{id}`。

分析：`GET /api/v1/analysis/overview`、`/signals`、`/patterns`、`/facets`、`/review`、`/themes`、`/emotions`、`/wordcloud`、`/mood`、`/sleep-contrast`、`/theme-pairs`、`/theme-timeline`、`/calendar`、`/phrases`、`/narratives`、`/api/v1/lexicon`，以及 `GET /api/v1/elements/suggestions`。

回看：`GET /api/v1/dreams/{id}/similar`、`GET /api/v1/dreams/{id}/reflection`；搜索：`GET /api/v1/dreams/search`，支持情绪、清晰度、睡眠、主题和元素筛选。

周报与导出：`GET /api/v1/reports/weekly`、`GET /api/v1/reports/history`、`GET /api/v1/reports/history/digests`、`POST /api/v1/reports/refresh`、`GET /api/v1/export/markdown`、`GET /api/v1/export/json`。

## 启动方式

环境要求：Go 1.22 或更高版本。

```bash
go build -o dream-workbench ./cmd/server
./dream-workbench -addr :8097 -data ./data/dreams.json
```

也可以直接运行：

```bash
go run ./cmd/server -addr :8097 -data ./data/dreams.json
```

配置项：`DREAM_ADDR`、`DREAM_DATA`、`DREAM_SECRET`、`DREAM_SESSION_TTL`、`DREAM_REFRESH_EVERY`、`DREAM_MAX_BODY`、`DREAM_SHUTDOWN_WINDOW`。默认数据文件是 `./data/dreams.json`，首次启动会自动创建。

## 端到端验收

```bash
GOCACHE=/tmp/dream-workbench-gocache go build ./...
GOCACHE=/tmp/dream-workbench-gocache go vet ./...
GOCACHE=/tmp/dream-workbench-gocache go test ./...
python3 /Users/tog_11/.codex/skills/go-create-new-project/scripts/project_guard.py verify \
  --scan-root /Users/tog_11/code/我的go \
  --project /Users/tog_11/code/我的go/30个/dream-narrative-analysis-workbench-117
```

启动服务后，在浏览器打开 `http://127.0.0.1:8097/`，注册用户，保存两条包含相同主题证据词的梦境，确认时间线和分析卡片更新，再打开模式、主题共现、相似梦境和反思页面，最后点击周报刷新和两个导出按钮。API 也可以用 Cookie 或 `Authorization: Bearer <token>` 调用。

## 设计边界

项目使用本地 JSON 文件作为默认存储，适合个人单实例运行；持久化写入采用临时文件替换，避免半写入文件。主题分析是基于证据词的透明规则，不对梦境做心理诊断。项目源代码不包含原始需求附件副本。

生成日期：2026-08-24。
