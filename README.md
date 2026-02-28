<div align="center">

# 🐍 灵宝市场

**高性能市场价格共享平台**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?style=flat&logo=next.js&logoColor=white)](https://nextjs.org)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis&logoColor=white)](https://redis.io)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker&logoColor=white)](https://docker.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[快速开始](#-快速开始) · [功能特性](#-功能特性) · [技术栈](#-技术栈) · [部署指南](#-部署指南) · [配置说明](#%EF%B8%8F-配置说明)

</div>

---

## ✨ 功能特性

- 🚀 **高并发架构** — Go + Fiber + Redis，轻松应对大量并发请求
- 🔄 **实时数据** — 灵宝代码即时生成、查询与价格共享
- 🌐 **国际化** — 中文 / English 双语支持
- 🔐 **管理后台** — 一键反馈处理与数据管理
- 🤖 **B 站自动导入** — 定时从 Bilibili 评论区抓取并导入代码数据
- ⏰ **自动清理** — 可配置的定时数据清理机制
- 🐳 **一键部署** — Docker Compose 开箱即用
- 🔁 **CI/CD** — GitHub Actions 自动构建并推送至 GHCR

## 🛠 技术栈

| 层级 | 技术 |
|:---:|:---|
| **后端** | Go 1.22 · Fiber v2 · Redis 7 · JWT · Viper |
| **前端** | Next.js 16 · React 19 · Tailwind CSS 4 · SWR · Framer Motion · NextAuth v5 · next-intl |
| **部署** | Docker Compose · Nginx · GitHub Actions · GHCR |

## 🚀 快速开始

### 前置要求

- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)

### 本地开发

```bash
# 克隆项目
git clone https://github.com/lmarch2/lingbao-market.git
cd lingbao-market

# 一键启动所有服务
docker-compose up -d --build
```

启动后访问：

| 服务 | 地址 |
|:---|:---|
| 🖥 前端界面 | http://localhost:3000 |
| ⚙️ 后端 API | http://localhost:8080 |
| 🌐 Nginx 网关 | http://localhost:8088 |

### 生产部署

```bash
# 1. 准备环境变量
cp .env.example .env
# 编辑 .env 填入实际配置

# 2. 使用生产配置启动
docker-compose -f docker-compose.prod.yml up -d
```

> 生产镜像从 `ghcr.io/lmarch2/lingbao-*` 拉取，无需本地构建。

## 📁 项目结构

```
lingbao-market/
├── backend/                # Go 后端服务
│   ├── cmd/               # 启动入口
│   ├── internal/
│   │   ├── api/           # HTTP 路由 & 处理器
│   │   ├── config/        # 配置加载
│   │   ├── model/         # 数据模型
│   │   └── service/       # 业务逻辑
│   ├── Dockerfile
│   └── go.mod
├── frontend/               # Next.js 前端
│   ├── app/               # App Router 页面
│   │   └── [locale]/      # 国际化路由
│   ├── components/        # React 组件
│   ├── i18n/              # 国际化配置
│   ├── messages/          # 翻译文件 (zh/en)
│   ├── lib/               # 工具库
│   ├── types/             # TypeScript 类型
│   ├── Dockerfile
│   └── package.json
├── deploy/                 # Nginx 配置
├── .github/workflows/      # CI/CD 流水线
├── docker-compose.yml      # 开发环境编排
├── docker-compose.prod.yml # 生产环境编排
└── .env.example            # 环境变量模板
```

## ⚙️ 配置说明

### 后端环境变量

| 变量 | 说明 | 默认值 |
|:---|:---|:---|
| `APP_ENV` | 运行环境 | `dev` |
| `REDIS_ADDR` | Redis 地址 | `redis:6379` |
| `ADMIN_USERNAME` | 管理员账号 | `admin` |
| `ADMIN_PASSWORD` | 管理员密码 | *必填* |
| `CLEANUP_TIME` | 每日清理时间 (24h) | `00:00` |
| `CLEANUP_TIMEZONE` | 时区 | `Local` |

<details>
<summary>📦 B 站自动导入配置</summary>

| 变量 | 说明 | 默认值 |
|:---|:---|:---|
| `BILIBILI_IMPORT_ENABLED` | 启用自动导入 | `false` |
| `BILIBILI_IMPORT_KEYWORD` | 搜索关键词 | `小马糕` |
| `BILIBILI_IMPORT_MIN_PRICE` | 最低导入价格 | `900` |
| `BILIBILI_IMPORT_LIMIT` | 每次导入上限 | `30` |
| `BILIBILI_IMPORT_SEARCH_PAGES` | 搜索页数 | `1` |
| `BILIBILI_IMPORT_SEARCH_PAGE_SIZE` | 每页视频数 | `20` |
| `BILIBILI_IMPORT_COMMENT_PAGES` | 评论抓取页数 | `1` |
| `BILIBILI_IMPORT_TIMEOUT_SECONDS` | 超时时间 (秒) | `60` |
| `BILIBILI_COOKIE` | 浏览器 Cookie (降低风控) | — |

</details>

### 前端环境变量

| 变量 | 说明 |
|:---|:---|
| `NEXT_PUBLIC_API_URL` | 后端 API 地址 |
| `AUTH_SECRET` | NextAuth 密钥 |
| `AUTH_URL` | 认证回调地址 |

## 🔧 常用命令

```bash
# 查看实时日志
docker-compose logs -f [service]

# 重启单个服务
docker-compose restart backend

# 停止所有服务
docker-compose down

# 清理数据（含 Redis 持久化数据）
docker-compose down -v
```

## 💡 使用说明

**用户端** — 访问首页生成 / 查询灵宝代码，查看对应价格

**管理端** — 登录管理后台处理用户反馈，管理数据

**自动清理** — 每天 `CLEANUP_TIME` 自动清除过期数据，清理后可自动从 B 站导入新数据

## 📄 License

[MIT](LICENSE)
