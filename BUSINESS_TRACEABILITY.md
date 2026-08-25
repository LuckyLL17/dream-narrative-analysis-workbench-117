# 业务追踪登记

本文件登记每个 Go 生产源文件的业务场景和调用链，确保新增模块都参与真实运行路径。项目没有测试文件或测试专用目录。

## 启动、配置与前端

| 文件 | 业务场景 | 生产调用链 |
|---|---|---|
| `cmd/server/main.go` | 组装服务、启动 HTTP 和后台调度 | 配置 -> Store -> Service -> HTTP -> Browser |
| `config/config.go` | 读取地址、数据文件、会话和调度参数 | main -> Config.Load |
| `web/assets.go` | 将浏览器资源嵌入单二进制 | main -> web.Handler -> HTTP |

## 领域与基础设施

| 文件 | 业务场景 | 生产调用链 |
|---|---|---|
| `internal/domain/enums.go` | 校验情绪、元素和状态 | DreamService -> ValidateDream |
| `internal/domain/models.go` | 表达用户、梦境、元素、分析和周报 | Store <-> Service <-> HTTP |
| `internal/domain/validation.go` | 守住记录和分析窗口不变量 | AuthService/DreamService -> domain.Validate |
| `internal/store/aggregates.go` | 读取主题、情绪和睡眠聚合数据 | AnalysisService -> Store -> JSON 文件 |
| `internal/store/dream_store.go` | 保存、更新和删除梦境记录 | DreamService -> DreamStore -> JSON 文件 |
| `internal/store/element_store.go` | 保存梦中元素及其与梦境的关联 | ElementService -> ElementStore -> JSON 文件 |
| `internal/store/filter.go` | 解析结构化筛选条件和分页边界 | DreamService -> Filter -> Query |
| `internal/store/job_store.go` | 持久化分析任务状态和失败信息 | JobQueue -> JobStore -> JSON 文件 |
| `internal/store/query.go` | 按时间、主题和情绪查询梦境 | DreamService -> Query -> JSON 文件 |
| `internal/store/report_query.go` | 查询窗口内报告和历史摘要 | ReportService -> ReportQuery -> JSON 文件 |
| `internal/store/report_store.go` | 写入和读取周报快照 | ReportService -> ReportStore -> JSON 文件 |
| `internal/store/store.go` | 统一数据文件读写和锁保护 | main -> Store -> JSON 文件 |
| `internal/store/user_store.go` | 保存用户账户和登录索引 | AuthService -> UserStore -> JSON 文件 |
| `internal/security/password.go` | 生成和验证密码摘要 | AuthService -> Password -> Session |
| `internal/security/session.go` | 维护会话生命周期和过期检查 | SessionService -> Session -> HTTP Middleware |
| `internal/security/token.go` | 签名和解析 Cookie 会话令牌 | Auth API -> Token -> Session Middleware |
| `pkg/clock/clock.go` | 周报周一到周日边界计算 | ReportService -> Clock |
| `pkg/ids/ids.go` | 生成用户、梦境、元素和任务 ID | Service/Queue -> IDs |
| `pkg/mathx/stats.go` | 平均值、比例、相关系数 | Analysis -> Math |
| `pkg/collections/top.go` | 主题和词频排序 | Analysis -> Collections |
| `pkg/jsonutil/json.go` | 统一蛇形 JSON 编解码，避免 API 与快照重复维护字段映射 | HTTP/Store/Export -> JSON codec |
| `pkg/textutil/normalize.go` | 叙事清洗和证据词匹配 | Text -> Textutil |

## 分析链路

| 文件 | 业务场景 | 生产调用链 |
|---|---|---|
| `internal/text/tokenize.go` | 中文词元提取 | Dream text -> Tokens |
| `internal/text/keywords.go` | 词频和情绪强度 | Analysis -> Keywords |
| `internal/text/elements.go` | 推荐七类梦中元素 | DreamService -> RecommendElements |
| `internal/text/themes.go` | 按证据词命中主题 | DreamService -> DetectThemes |
| `internal/text/scoring.go` | 搜索排序和情绪桶 | Timeline/Analysis -> Scoring |
| `internal/text/lexicon.go` | 维护主题证据词和情绪词典 | Themes/Keywords -> Lexicon |
| `internal/text/narrative.go` | 识别开端、转折和收束叙事段落 | PatternAnalysis -> Narrative |
| `internal/text/phrases.go` | 提取跨梦境重复短语 | PatternAnalysis -> PhraseMiner |
| `internal/analysis/window.go` | 时间窗口解析和边界归一化 | HTTP -> Analysis Window |
| `internal/analysis/overview.go` | 总览、主题和词频统计 | Overview API -> OverviewEngine |
| `internal/analysis/emotions.go` | 情绪时间序列和睡眠切片 | Emotions API -> EmotionEngine |
| `internal/analysis/insights.go` | 生成可解释观察提示 | OverviewEngine -> Insights |
| `internal/analysis/wordcloud.go` | 词云数据和叙事词集合 | Wordcloud API -> WordCloud |
| `internal/analysis/report.go` | 生成本周回顾和建议 | ReportService -> BuildReport |
| `internal/analysis/calendar.go` | 按日期压缩梦境密度和情绪 | AnalysisService -> CalendarAnalysis |
| `internal/analysis/cooccurrence.go` | 计算主题共现关系 | AnalysisService -> CooccurrenceAnalysis |
| `internal/analysis/facets.go` | 组织主题、元素和情绪筛选面 | AnalysisService -> FacetAnalysis |
| `internal/analysis/mood.go` | 比较情绪分布与叙事强度 | AnalysisService -> MoodAnalysis |
| `internal/analysis/patterns.go` | 识别重复短语和叙事模式 | AnalysisService -> PatternAnalysis |
| `internal/analysis/reflection.go` | 生成单条梦境反思提示 | AnalysisService -> ReflectionAnalysis |
| `internal/analysis/similarity.go` | 计算相似梦境候选 | AnalysisService -> SimilarityAnalysis |
| `internal/analysis/sleep.go` | 计算睡眠时长与梦境线索的关联 | AnalysisService -> SleepAnalysis |

## 应用服务与 API

| 文件 | 业务场景 | 生产调用链 |
|---|---|---|
| `internal/service/auth.go` | 注册、登录和用户恢复 | Auth API -> AuthService -> Store/Security |
| `internal/service/session.go` | 请求会话恢复 | Middleware -> SessionService |
| `internal/service/dream.go` | 创建、更新、列表、详情和删除 | Dream API -> DreamService -> Store/Text/Element |
| `internal/service/element.go` | 元素计数和推荐合并 | DreamService/Element API -> ElementService |
| `internal/service/analysis.go` | 聚合分析能力 | Analysis API -> AnalysisService |
| `internal/service/report.go` | 周报刷新、当前报告和历史 | Report API -> ReportService |
| `internal/service/export.go` | Markdown 和 JSON 个人导出 | Export API -> ExportService |
| `internal/service/health.go` | 服务和数据规模状态 | Health API -> HealthService |
| `internal/httpapi/analysis_handler.go` | 提供总览、模式、相似度和反思接口 | Browser -> AnalysisHandler -> AnalysisService |
| `internal/httpapi/app.go` | 组装路由依赖和 HTTP 应用 | main -> App -> Routes |
| `internal/httpapi/auth_handler.go` | 处理注册、登录和登出请求 | Browser -> AuthHandler -> AuthService |
| `internal/httpapi/decode.go` | 解码 JSON 请求并限制输入 | HTTP Handler -> Decode -> Domain |
| `internal/httpapi/dream_handler.go` | 处理梦境增删改查和时间线 | Browser -> DreamHandler -> DreamService |
| `internal/httpapi/element_handler.go` | 提供元素目录和推荐接口 | Browser -> ElementHandler -> ElementService |
| `internal/httpapi/export_handler.go` | 导出个人梦境 Markdown 和 JSON | Browser -> ExportHandler -> ExportService |
| `internal/httpapi/health_handler.go` | 返回服务健康和数据规模 | Browser -> HealthHandler -> HealthService |
| `internal/httpapi/middleware.go` | 恢复会话、记录指标和保护接口 | HTTP -> Middleware -> Session/Metrics |
| `internal/httpapi/report_handler.go` | 生成、读取和导出周报 | Browser -> ReportHandler -> ReportService |
| `internal/httpapi/response.go` | 统一 JSON 错误和成功响应 | Handler -> Response -> Browser |
| `internal/httpapi/routes.go` | 注册公开、认证和受保护路由 | App -> Routes -> Handlers |

## 后台和观测

| 文件 | 业务场景 | 生产调用链 |
|---|---|---|
| `internal/jobs/queue.go` | 排队和领取分析任务 | Dream changes -> Queue |
| `internal/jobs/refresh.go` | 执行周报刷新并记录失败 | Scheduler -> Refresher -> ReportService |
| `internal/jobs/scheduler.go` | 周期性触发后台任务 | main -> Scheduler |
| `internal/jobs/retry.go` | 失败任务退避重排 | maintenance -> RetryFailed |
| `internal/telemetry/logger.go` | 生产日志事件 | main/API/jobs -> Logger |
| `internal/telemetry/metrics.go` | 请求和错误计数 | main -> Metrics middleware |

## 规模说明

目标规模是有效 Go 生产代码超过 5000 行，但本项目以功能完整性为先。最终交付前会使用 `project_guard.py verify` 的有效代码扫描结果作为唯一统计，不会把测试、生成物、注释或机械重复代码计入，也不会为了达到数字加入不可达代码。
