# Build the landing for GitHub Pages and publish it to the `gh-pages` branch.
# Usage:  pwsh ./deploy.ps1
$ErrorActionPreference = "Stop"

$root = $PSScriptRoot
Set-Location "$root/frontend"

# Build with the Pages subpath and no backend (static fallback links).
if (Test-Path .env) { Rename-Item .env .env.bak -Force }
try {
  $env:PUBLIC_BASE = "/dmgr-web/"
  $env:PUBLIC_SITE = "https://khinmmad.github.io"
  npm run build
}
finally {
  if (Test-Path .env.bak) { Rename-Item .env.bak .env -Force }
  Remove-Item Env:\PUBLIC_BASE, Env:\PUBLIC_SITE -ErrorAction SilentlyContinue
}

# Publish dist/ to the gh-pages branch.
Set-Location dist
if (-not (Test-Path .nojekyll)) { New-Item -ItemType File .nojekyll | Out-Null }
if (-not (Test-Path .git)) {
  git init
  git remote add origin https://github.com/Khinmmad/dmgr-web.git
}
git add -A
git commit -m "deploy: $(Get-Date -Format s)"
git branch -M gh-pages
git push -f origin gh-pages

Write-Host "`nDeployed -> https://khinmmad.github.io/dmgr-web/" -ForegroundColor Green
