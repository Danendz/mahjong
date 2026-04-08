# Wuhan Mahjong (武汉麻将)

A real-time multiplayer Wuhan Mahjong game, also known as 红中赖子杠 (Red Dragon + Wild Card Kong). Play with friends online or practice against AI bots with configurable difficulty levels.

## Features

- **Real-time multiplayer** — 4-player games via WebSocket connections
- **AI bots** — Easy, medium, and hard difficulty levels
- **Guest mode** — No login required, jump straight into a game
- **Room-based gameplay** — Create rooms and share codes to invite friends
- **Full Wuhan mahjong rules** — Laizi (wild cards), rob kong (抢杠胡), kong draw win (杠上开花), last tile win (海底捞月), 258将 pair rule, and more
- **Scoring system** — Multiplicative multipliers with configurable score caps
- **Reconnection support** — Rejoin games seamlessly after disconnection

## Tech Stack

### Frontend

- **Vue 3** with Composition API
- **TypeScript**
- **Vite** — build tool and dev server
- **Pinia** — state management
- **Vue Router** — client-side routing
- **SCSS** — styling
- **Motion-v** — animations

### Backend

- **Go**
- **Gorilla WebSocket** — real-time communication
- **PostgreSQL** — database (optional for guest mode)
- **pgx** — PostgreSQL driver
- **Air** — hot reload during development

### Infrastructure

- **Docker & Docker Compose** — containerized backend and database
- **Task** — task runner for project automation
- **pnpm** — frontend package manager
- **JSON Schema** — shared type definitions with code generation for Go and TypeScript

## Prerequisites

- [Go](https://go.dev/) 1.26+
- [Node.js](https://nodejs.org/) with [pnpm](https://pnpm.io/)
- [Docker](https://www.docker.com/) & Docker Compose
- [Task](https://taskfile.dev/) (optional, for task runner commands)

## Getting Started

### Quick Start

Run both frontend and backend together:

```bash
task dev
```

### Manual Setup

**1. Start the database and backend:**

```bash
docker compose up --build
```

**2. Start the frontend dev server:**

```bash
cd frontend
pnpm install
pnpm dev
```

The frontend runs at `http://localhost:5173` and proxies API/WebSocket requests to the backend at `http://localhost:8080`.

### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `PORT` | Backend server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | — |
| `CORS_ORIGIN` | Allowed CORS origin | `*` |

See `.env.example` files in `frontend/` and `backend/` for more details.

## Available Tasks

| Command | Description |
|---|---|
| `task dev` | Start full development environment |
| `task build` | Build frontend and backend |
| `task db:up` | Start PostgreSQL database |
| `task db:down` | Stop database |
| `task db:reset` | Destroy and recreate database |
| `task generate` | Generate Go & TypeScript types from JSON Schema |
| `task test:backend` | Run backend tests |
| `task lint:frontend` | Lint frontend code |
| `task lint:backend` | Lint backend code |

## Project Structure

```
mahjong/
├── frontend/              # Vue 3 + TypeScript frontend
│   ├── src/
│   │   ├── components/    # UI components (tiles, overlays, action bar)
│   │   ├── composables/   # Vue composables (WebSocket, game connection)
│   │   ├── stores/        # Pinia stores (game, room, user state)
│   │   ├── types/         # TypeScript types (generated from schema)
│   │   ├── views/         # Page views (Lobby, Room, Game)
│   │   └── styles/        # Global SCSS styles
│   └── vite.config.ts
│
├── backend/               # Go backend
│   ├── cmd/server/        # Server entry point
│   └── internal/
│       ├── engine/        # Game engine (tiles, hand validation, scoring)
│       ├── room/          # Room management and game coordination
│       ├── ws/            # WebSocket handler
│       ├── bot/           # AI bot logic
│       └── db/            # Database layer and migrations
│
├── schema/                # JSON Schema definitions (shared types)
├── scripts/               # Utility scripts (type generation)
├── docs/                  # Game design documentation
├── docker-compose.yml
└── Taskfile.yml
```
