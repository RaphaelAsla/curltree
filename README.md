# Curltree 🌳

A terminal-based Linktree alternative that combines the power of SSH authentication with a beautiful TUI (Terminal User Interface) for managing your profile and links. Access profiles via `curl` for a true command-line experience!

## 🚀 Features

- **SSH Authentication**: Secure authentication using SSH public keys
- **Terminal UI**: Beautiful, interactive TUI built with Bubbletea
- **Public Profiles**: Access profiles via `curl curltree.dev/<username>`
- **Rate Limited**: Built-in rate limiting to prevent abuse
- **Self-Hosted**: Run on your own infrastructure
- **Database Agnostic**: Supports SQLite (dev) and PostgreSQL (production)
- **Configuration**: Flexible configuration via environment variables or JSON

## 📋 Quick Start

### Prerequisites

- Go 1.21+ (for building)
- SQLite or PostgreSQL
- SSH client for accessing the TUI

### Installation

1. **Clone and build:**
```bash
git clone <repository-url>
cd curltree
make build
```

2. **Generate SSH host key:**
```bash
make generate-keys
```

3. **Start the services:**
```bash
# Terminal 1: Start HTTP server
make run-server

# Terminal 2: Start SSH TUI server  
make run-tui
```

4. **Connect via SSH:**
```bash
ssh -p 23234 localhost
```

### Docker Installation

```bash
# Build and run with Docker
make docker-run

# Or use docker-compose
docker-compose up -d
```

## 🖥️ Usage

### Creating Your Profile

1. Connect to the TUI via SSH:
   ```bash
   ssh -p 23234 your-server.com
   ```

2. If you're a new user, you'll be prompted to create a profile:
   - Enter your full name
   - Choose a unique username
   - Add an "about" section
   - Add your links (websites, social media, etc.)

### Managing Your Profile

Once you have a profile, you can:

- **View**: See your profile in read-only mode
- **Edit** (`Ctrl+E`): Modify your information
- **Add Links** (`Ctrl+N`): Add new links while editing
- **Delete Links** (`Ctrl+D`): Remove the currently focused link
- **Delete Profile** (`Ctrl+D` on view mode): Permanently delete your profile

### Accessing Profiles

Profiles are publicly accessible via HTTP:

```bash
# Plain text format (perfect for terminals)
curl curltree.dev/username

# JSON format  
curl -H "Accept: application/json" curltree.dev/username
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CONFIG_PATH` | - | Path to JSON config file |
| `SERVER_HOST` | `localhost` | HTTP server host |
| `SERVER_PORT` | `8080` | HTTP server port |
| `SSH_HOST` | `localhost` | SSH server host |
| `SSH_PORT` | `23234` | SSH server port |
| `HOST_KEY_PATH` | `.ssh/curltree_host_key` | SSH host key path |
| `DB_TYPE` | `sqlite` | Database type (sqlite/postgres) |
| `DB_PATH` | `./curltree.db` | SQLite database path |
| `LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |

### Configuration File

Create a `config.json` file (see `config.example.json`):

```json
{
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "rate_limit": {
      "requests_per_minute": 60,
      "burst": 10
    }
  },
  "ssh": {
    "host": "0.0.0.0", 
    "port": 23234,
    "host_key_path": ".ssh/curltree_host_key"
  },
  "database": {
    "type": "sqlite",
    "path": "./curltree.db"
  },
  "logging": {
    "level": "info",
    "format": "text",
    "output": "stdout"
  }
}
```

## 🏗️ Architecture

### Components

1. **HTTP Server** (`cmd/server`): Serves public profiles via REST API
2. **TUI Server** (`cmd/tui`): SSH server with interactive terminal interface
3. **Database Layer** (`internal/database`): Data persistence with CRUD operations
4. **Authentication** (`internal/auth`): SSH public key based authentication
5. **Configuration** (`internal/config`): Flexible configuration management

### Database Schema

```sql
-- Users table
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    ssh_public_key TEXT UNIQUE,
    full_name TEXT,
    username TEXT UNIQUE,
    about TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- Links table  
CREATE TABLE links (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    name TEXT,
    url TEXT,
    position INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

## 🧪 Development

### Running Tests

```bash
make test

# With coverage
make test-coverage
```

### Development Mode

```bash
# Auto-reload server
make dev-server

# Auto-reload TUI
make dev-tui
```

### Code Quality

```bash
# Run all checks
make check

# Individual checks
make fmt      # Format code
make vet      # Run go vet
make lint     # Run linter (requires golangci-lint)
```

## 🚢 Deployment

### Production Deployment

1. **Configure for production:**
   ```bash
   export CONFIG_PATH=/path/to/production-config.json
   export DB_TYPE=postgres
   export DB_HOST=your-postgres-host
   # ... other production settings
   ```

2. **Build and deploy:**
   ```bash
   make build
   # Deploy binaries to your server
   ```

3. **Set up reverse proxy** (nginx example):
   ```nginx
   server {
       server_name curltree.yourdomain.com;
       location / {
           proxy_pass http://localhost:8080;
       }
   }
   ```

### Docker Deployment

```yaml
version: '3.8'
services:
  curltree:
    build: .
    ports:
      - "8080:8080"
      - "23234:23234"
    environment:
      - DB_PATH=/app/data/curltree.db
    volumes:
      - ./data:/app/data
```

## 📊 Monitoring

The application provides structured logging with the following information:

- HTTP requests (method, path, status code, duration)
- Rate limiting events
- SSH connections
- Profile operations (create, update, delete)
- Database operations

Configure logging output to integrate with your monitoring stack.

## 🔒 Security

- SSH public key authentication
- Input validation and sanitization
- SQL injection prevention
- Rate limiting on public endpoints
- No sensitive data in logs
- Secure session handling

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run `make check`
6. Submit a pull request

## 📝 License

[License information]

## 🙋‍♀️ Support

- [GitHub Issues](link-to-issues)
- [Documentation](link-to-docs)
- [Community Discord/Forum](link-to-community)

---

Made with ❤️ for the terminal enthusiasts!# curltree
