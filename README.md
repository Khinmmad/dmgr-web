# dmgr-web

Promotional landing page for **dmgr** — a modern device manager for Linux, Windows & macOS.

- **Frontend:** [Astro](https://astro.build) + [Tailwind CSS v4](https://tailwindcss.com) — static, bilingual (EN/ES), Catppuccin dark theme, scroll animations.
- **Backend:** Go microservice that serves download links from the GitHub Releases API (cached) and counts download clicks per platform.

The frontend works **fully static** on its own (bundled fallback links). The Go backend adds live release data + per-platform download counters.

```
dmgr-web/
├── frontend/      # Astro + Tailwind site
└── backend/       # Go releases API + download counter
```

## Frontend

```bash
cd frontend
npm install
npm run dev        # http://localhost:4321
npm run build      # → dist/  (deploy to Netlify / Vercel / GitHub Pages / any static host)
npm run preview    # serve the production build locally
```

Configure the backend URL in `frontend/.env`:

```
PUBLIC_API_BASE=http://localhost:8080
```

Leave it **empty** for a purely static deploy (no backend): download buttons use the
bundled GitHub release URLs and the download counters are hidden.

## Backend (Go)

```bash
cd backend
go run .                 # http://localhost:8080
# or build a binary:
go build -o dmgr-web-backend .
./dmgr-web-backend
```

Endpoints:

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/releases` | Latest release version + per-platform download info (cached 10 min). |
| `GET` | `/api/download/{platform}` | Increments the counter and 302-redirects to the asset. |
| `GET` | `/api/stats` | Download counts per platform (JSON). |
| `GET` | `/healthz` | Liveness check. |

`platform` is one of: `windows`, `macos`, `arch`, `debian`, `fedora`, `appimage`.
Assets are classified from the GitHub release by file extension, so the site
auto-updates when Windows (`.msi`) and macOS (`.dmg`) builds are published.

Environment variables:

| Var | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | Listen port. |
| `DATA_DIR` | `.` | Where `counts.json` is stored. |
| `GITHUB_TOKEN` | — | Optional; raises the GitHub API rate limit. |

Counters persist to `counts.json`. CORS is open (`*`) so the static site can call it
from anywhere.

## Deploy notes

- **Frontend:** any static host. Set `PUBLIC_API_BASE` to your deployed backend URL
  (or leave empty for static-only).
- **Backend:** any host that runs a Go binary (Fly.io, Railway, a small VPS, a
  container). Mount a volume for `counts.json` if you want counts to survive restarts.
