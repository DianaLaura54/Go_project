# 🍅📝 Pomodoro + Notes API — Go Project

A single Go project with two standalone programs:

| Binary | Purpose |
|---|---|
| `bin/pomodoro` | CLI Pomodoro timer with desktop notifications |
| `bin/server` | REST API for notes/todos with JWT-style auth |

---

## Project Structure

```
.
├── cmd/
│   ├── pomodoro/main.go   # Pomodoro CLI entry point
│   └── server/main.go     # REST API entry point
├── internal/
│   ├── auth/              # User registration, login, token creation/validation
│   ├── notes/             # Note/todo CRUD store
│   └── notify/            # Cross-platform desktop notification helper
├── Makefile
└── go.mod
```

---

## Quick Start

```bash
# Build both binaries
make build

# Run tests
make test
```

---

## 🍅 Pomodoro Timer

```bash
./bin/pomodoro [flags]

Flags:
  --work      Work session duration in minutes  (default: 25)
  --short     Short break in minutes            (default: 5)
  --long      Long break in minutes             (default: 15)
  --rounds    Rounds before a long break        (default: 4)
  --sessions  Total pomodoro sessions to run    (default: 4)
```

**Examples:**

```bash
# Default 4 × 25-minute sessions
./bin/pomodoro

# Custom: 2 × 50-minute sessions, 10-min breaks
./bin/pomodoro --work 50 --short 10 --sessions 2

# Quick test: 3 × 1-minute sessions
./bin/pomodoro --work 1 --short 1 --long 2 --sessions 3
```

Desktop notifications are sent at the start and end of each session/break. Supported on macOS (osascript), Linux (notify-send), and Windows (PowerShell). Falls back to stdout if unavailable.

---

## 📝 Notes/Todo REST API

```bash
./bin/server [--port 8080]
```

### Authentication

All `/notes` endpoints require a `Bearer` token in the `Authorization` header. Obtain a token by registering or logging in.

### Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | — | Health check |
| `POST` | `/auth/register` | — | Create account |
| `POST` | `/auth/login` | — | Get token |
| `GET` | `/notes` | ✅ | List all notes |
| `POST` | `/notes` | ✅ | Create a note |
| `GET` | `/notes/:id` | ✅ | Get a note |
| `PUT` | `/notes/:id` | ✅ | Update a note |
| `DELETE` | `/notes/:id` | ✅ | Delete a note |

### Example cURL flow

```bash
BASE=http://localhost:8080

# 1. Register
curl -s -X POST $BASE/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}'

# 2. Login (get TOKEN)
TOKEN=$(curl -s -X POST $BASE/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"secret123"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 3. Create a note
curl -s -X POST $BASE/notes \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"title":"Buy groceries","body":"Milk, eggs, bread","priority":"high","tags":["errands"]}'

# 4. List notes
curl -s $BASE/notes -H "Authorization: Bearer $TOKEN"

# 5. Update (mark done)
curl -s -X PUT $BASE/notes/note_1 \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"done":true}'

# 6. Delete
curl -s -X DELETE $BASE/notes/note_1 \
  -H "Authorization: Bearer $TOKEN"
```

### Note schema

```json
{
  "id":         "note_1",
  "user_id":    "abc123",
  "title":      "Buy groceries",
  "body":       "Milk, eggs, bread",
  "done":       false,
  "priority":   "high",
  "tags":       ["errands"],
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

Priority values: `low`, `medium`, `high`

---

## Notes

- User data and notes are stored **in-memory**. Restart the server and they reset. For persistence, swap `UserStore` / `Store` for a SQLite or PostgreSQL backend.
- Tokens expire after **24 hours**.
- The token secret defaults to `"super-secret-change-me"` — override it in `cmd/server/main.go` or expose it as an env var before going to production.
