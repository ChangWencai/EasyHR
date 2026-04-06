<!-- GSD:project-start source:PROJECT.md -->
## Project

**易人事（EasyHR）**

为小微企业/个体户老板（10-50人规模）打造的轻量化、一站式人事管理APP。解决无专职HR的小老板在员工入职、社保、工资、个税、财务记账等基础人事事务中的操作痛点。核心体验：3步内完成核心操作，零学习成本。

产品包含：老板端原生APP（Android/iOS）、H5管理后台（Vue 3）、员工端微信小程序。

**Core Value:** **简单、好用、省时间** — 老板打开APP第一时间知道要做什么，3步完成核心人事操作，无需专业知识。

### Constraints

- **操作体验**: 核心功能操作步骤 ≤ 3步 — 产品核心差异化，必须坚持
- **性能**: API 响应 ≤ 500ms（95%请求），APP端操作响应 ≤ 2s — 小微企业老板不会等待
- **并发**: 支持 ≥ 1000 同时在线用户，单企业 ≤ 50人同时操作无卡顿
- **可用性**: 全年 ≥ 99.9%，7×24小时稳定运行
- **兼容性**: Android 8.0+ / iOS 12.0+
- **合规**: 符合《个人信息保护法》《劳动合同法》等法律法规
- **成本**: V1.0 核心功能完全免费，技术成本需可控
- **部署**: Docker + 阿里云 ECS，单二进制部署
<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->
## Technology Stack

## Recommended Stack
### Backend Core (Go)
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Go | 1.23+ | 编程语言 | 高性能、编译为单二进制、并发模型优秀、中国Go生态成熟 | HIGH |
| Gin | v1.12.0 | HTTP框架 | 中国最流行的Go Web框架、社区活跃、中间件生态丰富、性能优秀。对比Echo(v5.1.0)路由语法更直观，对比Fiber(v3.1.0)不依赖fasthttp因此标准库兼容性更好 | HIGH |
| GORM | v1.31.1 | ORM | Go生态最成熟的ORM、支持PostgreSQL全部特性（JSONB、事务、迁移）、中文文档完善、Auto Migration适合快速迭代。Ent(v0.14.6)代码生成更类型安全但学习曲线陡峭，不适合小团队快速迭代 | HIGH |
| PostgreSQL | 16+ | 主数据库 | ACID事务保障、JSONB支持灵活存储（社保政策库）、优秀的中国时区支持、Row Level Security可增强多租户隔离 | HIGH |
| go-redis | v9.18.0 | Redis客户端 | 官方推荐客户端、支持Redis 7全部特性、连接池管理成熟、集群模式支持完善 | HIGH |
### Backend Infrastructure
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Viper | v1.21.0 | 配置管理 | Go生态事实标准的配置库、支持多格式(YAML/TOML/ENV)、环境变量覆盖适合Docker部署 | HIGH |
| Zap | v1.27.1 | 结构化日志 | Uber出品、性能极高（比标准库快10x+）、结构化日志便于问题排查、支持日志分级 | HIGH |
| golang-jwt | v5.3.1 | JWT认证 | golang-jwt官方维护（原dgrijalva已不维护）、v5 API简洁安全、支持RS256/HS256 | HIGH |
| go-playground/validator | v10.30.2 | 参数校验 | Gin生态标配、struct tag声明式校验、内置丰富验证规则（手机号、身份证可自定义）、中文错误消息支持 | HIGH |
| golang-migrate | v4.19.1 | 数据库迁移 | Go生态主流迁移工具、支持CLI和代码两种使用方式、PostgreSQL原生支持、迁移文件版本化管理 | HIGH |
### Backend Business Libraries
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| excelize | v2.10.1 | Excel读写 | Go生态最强大的Excel库、支持xlsx全部特性、样式/图表/公式、工资条导出和考勤导入必备 | HIGH |
| go-pdf/fpdf | v0.9.0 | PDF生成 | 纯Go实现、无CGO依赖、适合生成合同PDF模板、轻量够用。如需复杂排版可后期引入wkhtmltopdf | MEDIUM |
| resty | v2.17.2 | HTTP客户端 | 简洁的RESTful API、重试/超时/中间件支持完善、比标准库代码量少50%+、用于调用微信API和短信服务 | HIGH |
| silenceper/wechat | v2.1.12 | 微信SDK | 覆盖小程序登录/公众号/微信支付、持续维护（2026-02更新）、API设计清晰 | MEDIUM |
| gopay | v1.5.117 | 支付SDK | 支持微信支付/支付宝、极其活跃的维护（2026-04更新）、微信小程序支付必备 | HIGH |
| asynq | v0.26.0 | 异步任务队列 | 基于Redis的可靠任务队列、支持定时任务/重试/优先级、比直接用cron更可靠、适合工资核算/社保提醒等异步场景 | HIGH |
| gocron | v2.19.1 | 定时任务 | 比robfig/cron（已停更4年+）更活跃、支持分布式锁、单机/分布式灵活切换、适合社保到期提醒等定时场景 | HIGH |
### Backend Security & Compliance
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| crypto (标准库) | - | AES-256-GCM加密 | 标准库crypto/aes+crypto/cipher、无需第三方依赖、满足个人信息保护法加密要求 | HIGH |
| crypto/sha256 (标准库) | - | 哈希索引 | 敏感字段（身份证号）用SHA-256生成哈希索引、标准库直接使用 | HIGH |
| casbin | 最新 | RBAC权限 | Go生态最成熟的权限框架、支持OWNER/ADMIN/MEMBER三级RBAC、策略热加载、与Gin集成成熟 | HIGH |
### Backend Data & Storage
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Aliyun OSS SDK | v3.0.2 | 文件存储 | 阿里云官方维护、合同/工资条/凭证文件存储、签名URL直传减少服务端压力 | HIGH |
| testify | v1.11.1 | 测试断言 | Go测试生态标配、assert/mock/suite三件套、提高测试可读性 | HIGH |
### Frontend - H5管理后台 (Vue 3)
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Vue | 3.5.32 | 前端框架 | 项目已确定、Composition API成熟稳定、中国前端生态首选 | HIGH |
| Element Plus | 2.13.6 | UI组件库 | 项目已确定、Vue 3生态最成熟的组件库、表单/表格/对话框开箱即用、后台管理标配 | HIGH |
| Vite | 8.0.3 | 构建工具 | Vue官方推荐、开发体验极快（HMR <100ms）、Rollup打包优化 | HIGH |
| TypeScript | 6.0.2 | 类型安全 | Vue 3 + TS是最佳实践、减少运行时错误、IDE支持完善 | HIGH |
| Pinia | 3.0.4 | 状态管理 | Vue官方推荐、比Vuex更轻量、TS支持更好、Composition API风格 | HIGH |
| Vue Router | 5.0.4 | 路由管理 | Vue官方路由、支持路由守卫（权限控制）、懒加载优化首屏 | HIGH |
| @vueuse/core | 14.2.1 | 组合式工具 | Vue版lodash、100+实用composable（useLocalStorage/useDebounce等）、减少手写样板代码 | HIGH |
| Axios | 1.14.0 | HTTP客户端 | 拦截器机制适合统一token管理/错误处理、请求取消/重试、中国前端标配 | HIGH |
| dayjs | 1.11.20 | 日期处理 | 极轻量（2KB vs moment 67KB）、链式API与moment兼容、中文locale支持、薪资月份计算必备 | HIGH |
| ECharts | 6.0.0 | 图表 | 百度出品、中国数据可视化标准、Vue集成方案成熟、后续报表功能需要 | HIGH |
| xlsx (SheetJS) | 0.18.5 | Excel导入导出 | 浏览器端Excel处理、支持工资表导入/导出、中文兼容性好 | MEDIUM |
### Frontend - H5 Dev Tools
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| ESLint | 10.2.0 | 代码检查 | Vue + TS项目标配、配合eslint-plugin-vue | HIGH |
| Prettier | 3.8.1 | 代码格式化 | 团队统一代码风格、保存时自动格式化 | HIGH |
| Sass | 1.99.0 | CSS预处理器 | Element Plus官方推荐、变量/mixin系统完善 | HIGH |
| unplugin-auto-import | 21.0.0 | 自动导入 | Vue/Vue Router/Pinia API自动导入、减少样板import语句 | HIGH |
| unplugin-vue-components | 32.0.0 | 组件自动导入 | Element Plus组件按需自动导入、减少打包体积 | HIGH |
### Mobile - Android (Kotlin)
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Kotlin | 2.x | 编程语言 | 项目已确定、Google官方推荐、与Java完全互操作、协程支持异步 | HIGH |
| Jetpack Compose | 最新稳定版 | UI框架 | Android现代UI框架、声明式UI开发效率高、Material 3支持完善 | HIGH |
| Material 3 | 最新 | 设计系统 | Google最新设计规范、Composable组件丰富、暗色模式支持 | HIGH |
| Ktor | 最新 | HTTP客户端 | Kotlin Multiplatform原生HTTP库、协程友好、比Retrofit更轻量现代 | MEDIUM |
| Coil | 最新 | 图片加载 | Kotlin Coroutines原生、Compose集成优秀、比Glide更现代 | HIGH |
| DataStore | 最新 | 本地存储 | SharedPreferences替代品、协程/Flow支持、类型安全 | HIGH |
| Hilt | 最新 | 依赖注入 | Google官方推荐、基于Dagger2简化版、编译时检查 | HIGH |
### Mobile - iOS (Swift)
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Swift | 5.9+ | 编程语言 | 项目已确定、性能优秀、安全性高 | HIGH |
| SwiftUI | 最新 | UI框架 | 声明式UI、与Jetpack Compose对称开发、iOS 14+支持覆盖目标用户 | HIGH |
| Alamofire | 最新 | HTTP客户端 | Swift生态最成熟的网络库、拦截器/重试/认证管理完善 | HIGH |
| Kingfisher | 最新 | 图片加载 | Swift生态图片加载标准、缓存管理优秀、SwiftUI支持 | HIGH |
| SwiftData | 最新 | 本地持久化 | Apple最新数据持久化框架、替代Core Data、Swift原生API | MEDIUM |
### WeChat Mini Program (员工端)
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| 原生微信小程序 | 基础库3.x+ | 开发框架 | **不用Taro/uni-app**。理由：员工端功能简单（查看工资条/合同/社保记录、提交报销），原生开发性能最优、API兼容性最好、无需跨端框架增加复杂度 | HIGH |
| WeUI | 最新 | UI组件库 | 微信官方设计规范、小程序原生组件体验、员工端无需定制UI | HIGH |
| wx.request | - | 网络请求 | 小程序原生API够用、简单封装即可，无需引入额外网络库 | HIGH |
| miniprogram-ci | 最新 | CI/CD上传 | 微信官方CI工具、自动化上传代码到微信后台 | HIGH |
### Infrastructure & Deployment
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Docker | 27+ | 容器化 | 单二进制+Docker部署、开发环境一致性、阿里云ECS上运行 | HIGH |
| Aliyun ECS | - | 服务器 | 项目已确定、国内访问延迟低、按量付费成本可控 | HIGH |
| Aliyun OSS | - | 对象存储 | 项目已确定、合同/工资条/凭证文件存储、CDN加速 | HIGH |
| Aliyun SMS | - | 短信服务 | 国内短信到达率最高、验证码发送、支持短信签名和模板 | HIGH |
| Nginx | 1.26+ | 反向代理 | SSL终结、静态资源服务、负载均衡（后续扩容）、Go服务前置 | HIGH |
## Alternatives Considered
| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| HTTP框架 | Gin | Echo v5 | Echo也不错但中国Gin生态更大、中文资源更多；Gin中间件更丰富 |
| HTTP框架 | Gin | Fiber v3 | Fiber依赖fasthttp而非标准net/http，标准库兼容性差，第三方中间件生态不如Gin |
| HTTP框架 | Gin | go-zero v1.10 | go-zero是微服务框架，对模块化单体过重，自带代码生成不适合定制化需求 |
| ORM | GORM | Ent | Ent代码生成方式学习曲线陡峭、对快速迭代不友好、小团队GORM更高效 |
| ORM | GORM | sqlx | sqlx手写SQL维护成本高、GORM的Auto Migration在V1.0快速迭代阶段更高效 |
| 状态管理 | Pinia | Vuex 5 | Vuex已进入维护模式、Pinia是Vue官方推荐的未来方案 |
| 小程序框架 | 原生 | Taro | 员工端功能极简（5-6个页面），跨端框架增加构建复杂度和调试成本，不值得 |
| 小程序框架 | 原生 | uni-app | 同上，uni-app更适合需要同时覆盖多端（H5/多端小程序/App）的场景 |
| 定时任务 | gocron | robfig/cron | robfig/cron自2021年起停更，gocron活跃维护且功能更丰富（分布式锁、web UI） |
| 日志 | Zap | logrus | logrus已进入维护模式、性能远不如Zap |
| PDF生成 | go-pdf/fpdf | wkhtmltopdf | wkhtmltopdf需要CGO+外部二进制依赖，Docker镜像增大200MB+，fpdf纯Go够用 |
| 缓存 | go-redis直接使用 | eko/gocache | gocache提供抽象层但增加复杂度，go-redis直接操作足够清晰 |
| 异步队列 | asynq | RabbitMQ/Kafka | V1.0不需要消息中间件的复杂度，asynq基于Redis足够可靠，运维成本低 |
## Go Module Dependencies (go.mod 核心依赖)
## Frontend Dependencies (package.json 核心依赖)
## API Documentation
| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| swag (swaggo) | v2.0.0-rc5 | API文档生成 | Go注释生成OpenAPI 2.0文档、与Gin集成(gin-swagger)、自动同步代码变更 | MEDIUM |
| OpenAPI/Swagger UI | - | API文档展示 | 前后端协作标准、支持在线测试API、微信小程序开发对照接口 | HIGH |
## Key Version Notes
## Sources
- GitHub Release API (gin v1.12.0, echo v5.1.0, fiber v3.1.0, gorm v1.31.1, ent v0.14.6, go-redis v9.18.0, viper v1.21.0, zap v1.27.1, jwt v5.3.1, validator v10.30.2, excelize v2.10.1, resty v2.17.2, silenceper/wechat v2.1.12, gopay v1.5.117, asynq v0.26.0, gocron v2.19.1, golang-migrate v4.19.1, testify v1.11.1) -- 2026-04-06
- npm Registry (vue 3.5.32, element-plus 2.13.6, pinia 3.0.4, vue-router 5.0.4, vite 8.0.3, axios 1.14.0, typescript 6.0.2, @vueuse/core 14.2.1) -- 2026-04-06
- robfig/cron GitHub (最后提交 2021-01-06, 确认停更) -- 2026-04-06
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd:quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd:debug` for investigation and bug fixing
- `/gsd:execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd:profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
