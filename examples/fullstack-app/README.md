# Full-Stack App Example

This example demonstrates using Ralph to build a full-stack task management application with React frontend and Go backend.

## Project Overview

A modern full-stack application featuring:
- **Frontend**: React + TypeScript + Vite + Tailwind CSS
- **Backend**: Go REST API with JWT authentication
- **Database**: PostgreSQL in Docker
- **Infrastructure**: Docker Compose for development and production

## Architecture

```
fullstack-app/
├── frontend/           # React application
│   ├── src/
│   │   ├── components/ # Reusable UI components
│   │   ├── pages/      # Route pages
│   │   └── context/    # React context (auth, etc.)
│   └── package.json
├── backend/            # Go API
│   ├── cmd/api/        # Entry point
│   ├── internal/
│   │   ├── auth/       # Authentication
│   │   ├── db/         # Database
│   │   └── handlers/   # HTTP handlers
│   └── go.mod
├── docker-compose.yml  # Development setup
└── plan.json           # Ralph plan file
```

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker and Docker Compose
- Ralph installed

### Using Ralph

1. **Check the plan:**
   ```bash
   ralph -status -plan plan.json
   ```

2. **View milestones:**
   ```bash
   ralph -milestones -plan plan.json
   ```

3. **Run iterations:**
   ```bash
   ralph -iterations 15 -plan plan.json -verbose
   ```

## Milestones

| Milestone | Description | Features |
|-----------|-------------|----------|
| **Setup** | Project infrastructure | Monorepo, backend API, frontend setup |
| **Auth** | User authentication | Database, auth API, login/register UI |
| **Core** | Main functionality | Task CRUD API, task management UI |
| **Quality** | Production readiness | Tests, Docker deployment |

## Configuration

Create `.ralph.yaml`:

```yaml
# Use pnpm for frontend (detected from monorepo)
build_system: auto

# Custom commands for full-stack
typecheck: make typecheck  # Runs both frontend and backend checks
test: make test            # Runs both frontend and backend tests

plan: plan.json
verbose: true
iterations: 10
```

## Running the App

After Ralph completes:

```bash
# Development mode
docker-compose up -d postgres  # Start database
cd backend && go run ./cmd/api &  # Start backend
cd frontend && npm run dev     # Start frontend

# Or production mode
docker-compose -f docker-compose.prod.yml up
```

## Features

### Authentication
- Email/password registration
- JWT-based authentication
- Protected routes
- Session management

### Task Management
- Create, read, update, delete tasks
- Mark tasks as complete
- User-specific task lists

## Multi-Agent Mode

For faster development, use multiple agents:

```bash
ralph -iterations 10 -multi-agent -agents agents.json
```

Create `agents.json`:
```json
{
  "agents": [
    {"id": "backend", "role": "implementer", "command": "cursor-agent", "specialization": "go backend"},
    {"id": "frontend", "role": "implementer", "command": "cursor-agent", "specialization": "react frontend"},
    {"id": "tester", "role": "tester", "command": "cursor-agent"}
  ],
  "max_parallel": 2
}
```
