# dmgr-web

Promotional landing page for **dmgr** — a modern device manager for Linux, Windows & macOS.

**🌐 Live:** https://khinmmad.github.io/dmgr-web/

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

## Deploy

The frontend is live on **GitHub Pages** at https://khinmmad.github.io/dmgr-web/
and deploys **automatically** on every push to `main` via GitHub Actions
(`.github/workflows/deploy.yml`). The workflow builds with `PUBLIC_BASE=/dmgr-web/`
and publishes to Pages — no manual step needed. You can also trigger it from the
**Actions** tab (*Run workflow*).

### Host the backend (for live download counters)

GitHub Pages can't run the Go service, so the public site uses the bundled fallback
links and hides the counters until a backend is hosted. Config is included:

**Fly.io:**
```bash
cd backend
fly launch --copy-config --now      # pick a unique app name; creates the volume
# → note the URL, e.g. https://dmgr-web-<you>.fly.dev
```

**Render.com:** push to GitHub, then *New → Blueprint* and select this repo
(`render.yaml`). **Docker (anywhere):**
```bash
cd backend && docker build -t dmgr-web-backend . && docker run -p 8080:8080 -v dmgr_data:/data dmgr-web-backend
```

Then point the frontend at it and redeploy:
```bash
# set PUBLIC_API_BASE=https://your-backend-url  in frontend/.env  (and in deploy.ps1)
pwsh ./deploy.ps1
```

### Custom domain

1. Put your domain in `frontend/public/CNAME` (e.g. `dmgr.app`).
2. Set `PUBLIC_BASE` to `/` (root) instead of `/dmgr-web/` when building, since a
   custom domain serves from the root. Update `deploy.ps1` accordingly.
3. Add the DNS records GitHub shows under *Settings → Pages → Custom domain*
   (an `A`/`ALIAS` to GitHub Pages, or a `CNAME` to `khinmmad.github.io`).
