# Lingbao Market 2026 (Snake Year Edition)

High-performance market price sharing platform built for speed and concurrency.

## Stack
- **Backend**: Go 1.22+, Fiber v2, Redis 7
- **Frontend**: Next.js 14, Tailwind CSS, SWR
- **Infrastructure**: Docker Compose

## Quick Start

```bash
# Start all services
docker-compose up -d --build
```

## Structure
- `/backend`: Go API service
- `/frontend`: Next.js web application
- `/deploy`: Nginx and deployment configs
