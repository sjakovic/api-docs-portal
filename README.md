# API Docs Portal

A self-hosted platform for managing API documentation with user authentication and role-based access control. Serve OpenAPI/Swagger, Markdown, and external docs from one place.

**One binary. Zero dependencies. Full control.**

---

## Features

- **Multi-format support** - OpenAPI/Swagger (with Swagger UI), Markdown, external URLs via iframe
- **User management** - create accounts, assign admin or viewer roles
- **Per-doc access control** - define exactly which users can see which docs
- **Single binary** - everything is embedded: templates, static files, Swagger UI
- **SQLite database** - no database server to install or manage
- **Configurable branding** - set your own site title from the admin panel
- **Interactive setup** - no default credentials, the first run walks you through creating your admin account

## Quick Start

### Option 1: Download Binary

Download the latest release for your platform from the [Releases](https://github.com/sjakovic/api-docs-portal/releases) page.

```bash
chmod +x api-docs-portal
export JWT_SECRET=$(openssl rand -hex 32)
./api-docs-portal
```

Open `http://localhost:8080` - the setup wizard will guide you through creating your admin account.

### Option 2: Docker

```bash
git clone https://github.com/sjakovic/api-docs-portal.git
cd api-docs-portal
cp .env.example .env
# Edit .env and set JWT_SECRET to a random string
docker compose up -d
```

### Option 3: Build from Source

Requires Go 1.22+.

```bash
git clone https://github.com/sjakovic/api-docs-portal.git
cd api-docs-portal
make swagger-ui
make build
export JWT_SECRET=$(openssl rand -hex 32)
./bin/api-docs-portal
```

## Usage

### 1. Initial Setup

On first run, you'll see a setup page where you:
- Set your site title (e.g., your company name)
- Create your admin account with your own email and password

### 2. Add Documentation

Go to **Manage Docs** and click **Add Doc**. Choose the type:

| Type | What to provide | How it's rendered |
|------|----------------|-------------------|
| **OpenAPI/Swagger** | Paste or upload a JSON/YAML spec | Swagger UI |
| **Markdown** | Write or paste Markdown content | Rendered HTML |
| **External URL** | Link to any existing docs page | iframe embed |

### 3. Create Users

Go to **Users** and create accounts for your team members. Each user gets a role:
- **Admin** - full access to all docs and the admin panel
- **Viewer** - can only see docs they've been granted access to

### 4. Assign Permissions

On the **Manage Docs** page, click **Permissions** on any doc to choose which users can view it.

### 5. Share

Send your team to your portal URL - they'll only see the docs they have access to.

## Configuration

All configuration is done through environment variables.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Server bind address |
| `DB_PATH` | `./portal.db` | Path to SQLite database file |
| `JWT_SECRET` | - | **Required.** Secret key for signing auth tokens |
| `JWT_EXPIRY` | `24h` | How long login sessions last |
| `BASE_URL` | - | Base URL when running behind a reverse proxy |

## Production Deployment

### Standalone Binary

The simplest deployment - copy the binary to your server and run it:

```bash
# On your server
export JWT_SECRET="your-secret-key"
export DB_PATH="/var/lib/api-docs-portal/portal.db"
./api-docs-portal
```

Consider using a process manager like `systemd` to keep it running:

```ini
[Unit]
Description=API Docs Portal
After=network.target

[Service]
ExecStart=/opt/api-docs-portal/api-docs-portal
Environment=JWT_SECRET=your-secret-key
Environment=DB_PATH=/var/lib/api-docs-portal/portal.db
Restart=always
User=www-data

[Install]
WantedBy=multi-user.target
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 443 ssl;
    server_name docs.example.com;

    ssl_certificate /etc/ssl/certs/docs.example.com.pem;
    ssl_certificate_key /etc/ssl/private/docs.example.com.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Docker

```bash
docker compose up -d
```

Data is persisted in a Docker volume. Set `JWT_SECRET` in your `.env` file.

## Development

```bash
make swagger-ui   # Download Swagger UI assets (first time only)
make dev           # Run in development mode
make build         # Build production binary
make test          # Run tests
```

## Tech Stack

- **Go** - single binary, cross-platform
- **Chi** - lightweight HTTP router
- **SQLite** - embedded database (via modernc.org/sqlite, pure Go, no CGO)
- **HTMX** - interactive UI without a JavaScript framework
- **Swagger UI** - OpenAPI spec rendering (embedded)
- **goldmark** - Markdown rendering

## License

MIT - see [LICENSE](LICENSE) for details.

## Author

Built by [Simo Jakovic](https://jakovic.com)
