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
(served from the `gh-pages` branch). To rebuild and publish:

```powershell
pwsh ./deploy.ps1
```

It builds with `PUBLIC_BASE=/dmgr-web/` (the Pages subpath) and no backend, then
force-pushes `frontend/dist/` to `gh-pages`. GitHub Pages serves it within ~1 min.

### Notes

- **Frontend:** any static host works. For Pages the subpath base is required; for a
  root domain (Netlify / Vercel / Cloudflare Pages) leave `PUBLIC_BASE` unset.
  Set `PUBLIC_API_BASE` to your deployed backend URL to enable live release data +
  download counters (otherwise the bundled fallback links are used).
- **Backend:** the Go service needs a host that runs a binary (Fly.io, Railway, a
  small VPS, a container) — GitHub Pages can't run it. Mount a volume for
  `counts.json` to persist counts. Once hosted, set `PUBLIC_API_BASE` and redeploy
  the frontend.
