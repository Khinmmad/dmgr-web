// dmgr-web-backend: serves download links from the GitHub Releases API (cached)
// and counts download clicks per platform. Standard library only.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	repo         = "Khinmmad/dmgr"
	aurURL       = "https://aur.archlinux.org/packages/dmgr-desktop"
	releasesPage = "https://github.com/" + repo + "/releases/latest"
	cacheTTL     = 10 * time.Minute
)

// platforms we surface, in display order.
var platforms = []string{"windows", "macos", "arch", "debian", "fedora", "appimage"}

// ── API response shapes ───────────────────────────────────────────────────────

type PlatformInfo struct {
	Available bool   `json:"available"`
	URL       string `json:"url,omitempty"`
	Filename  string `json:"filename,omitempty"`
	Command   string `json:"command,omitempty"`
	Size      int64  `json:"size,omitempty"`
}

type ReleasesResp struct {
	Version     string                  `json:"version"`
	PublishedAt string                  `json:"published_at,omitempty"`
	Platforms   map[string]PlatformInfo `json:"platforms"`
}

// ── GitHub API shapes ─────────────────────────────────────────────────────────

type ghRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt string    `json:"published_at"`
	Assets      []ghAsset `json:"assets"`
}
type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// ── Release cache ─────────────────────────────────────────────────────────────

type cache struct {
	mu      sync.Mutex
	data    *ReleasesResp
	fetched time.Time
}

func (c *cache) get() *ReleasesResp {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data != nil && time.Since(c.fetched) < cacheTTL {
		return c.data
	}
	fresh, err := fetchReleases()
	if err != nil {
		log.Printf("releases fetch failed: %v", err)
		if c.data != nil {
			return c.data // serve stale on error
		}
		return fallback()
	}
	c.data = fresh
	c.fetched = time.Now()
	return fresh
}

func fetchReleases() (*ReleasesResp, error) {
	req, _ := http.NewRequest("GET",
		"https://api.github.com/repos/"+repo+"/releases/latest", nil)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "dmgr-web-backend")
	if tok := os.Getenv("GITHUB_TOKEN"); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, &httpError{resp.StatusCode}
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return mapRelease(&rel), nil
}

type httpError struct{ code int }

func (e *httpError) Error() string { return "github status " + http.StatusText(e.code) }

// mapRelease classifies assets into platforms by filename.
func mapRelease(rel *ghRelease) *ReleasesResp {
	out := &ReleasesResp{
		Version:     rel.TagName,
		PublishedAt: rel.PublishedAt,
		Platforms:   map[string]PlatformInfo{},
	}
	for _, p := range platforms {
		out.Platforms[p] = PlatformInfo{Available: false}
	}
	// Arch is always available via the AUR (not a release asset).
	out.Platforms["arch"] = PlatformInfo{Available: true, URL: aurURL, Command: "paru -S dmgr-desktop"}

	for _, a := range rel.Assets {
		name := strings.ToLower(a.Name)
		var plat string
		switch {
		case strings.HasSuffix(name, ".msi"), strings.HasSuffix(name, "-setup.exe"), strings.HasSuffix(name, ".exe"):
			plat = "windows"
		case strings.HasSuffix(name, ".dmg"), strings.HasSuffix(name, ".app.tar.gz"):
			plat = "macos"
		case strings.HasSuffix(name, ".deb"):
			plat = "debian"
		case strings.HasSuffix(name, ".rpm"):
			plat = "fedora"
		case strings.HasSuffix(name, ".appimage"):
			plat = "appimage"
		default:
			continue
		}
		// Keep the first match per platform.
		if info := out.Platforms[plat]; !info.Available {
			out.Platforms[plat] = PlatformInfo{
				Available: true,
				URL:       a.BrowserDownloadURL,
				Filename:  a.Name,
				Size:      a.Size,
			}
		}
	}
	return out
}

// fallback used only if GitHub is unreachable on first call.
func fallback() *ReleasesResp {
	return &ReleasesResp{
		Version: "",
		Platforms: map[string]PlatformInfo{
			"windows":  {Available: false},
			"macos":    {Available: false},
			"arch":     {Available: true, URL: aurURL, Command: "paru -S dmgr-desktop"},
			"debian":   {Available: false},
			"fedora":   {Available: false},
			"appimage": {Available: false},
		},
	}
}

// ── Download counters ─────────────────────────────────────────────────────────

type counters struct {
	mu    sync.Mutex
	path  string
	Count map[string]int
}

func newCounters(path string) *counters {
	c := &counters{path: path, Count: map[string]int{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &c.Count)
	}
	return c
}

func (c *counters) inc(platform string) {
	c.mu.Lock()
	c.Count[platform]++
	snapshot, _ := json.Marshal(c.Count)
	c.mu.Unlock()
	if err := os.WriteFile(c.path, snapshot, 0o644); err != nil {
		log.Printf("could not persist counts: %v", err)
	}
}

func (c *counters) snapshot() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int, len(c.Count))
	for k, v := range c.Count {
		out[k] = v
	}
	return out
}

// ── HTTP ──────────────────────────────────────────────────────────────────────

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "."
	}
	counts := newCounters(filepath.Join(dataDir, "counts.json"))
	rc := &cache{}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/releases", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, rc.get())
	}))

	mux.HandleFunc("GET /api/stats", cors(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, counts.snapshot())
	}))

	mux.HandleFunc("GET /api/download/{platform}", cors(func(w http.ResponseWriter, r *http.Request) {
		plat := r.PathValue("platform")
		info, ok := rc.get().Platforms[plat]
		if !ok || !info.Available || info.URL == "" {
			http.Redirect(w, r, releasesPage, http.StatusFound)
			return
		}
		counts.inc(plat)
		http.Redirect(w, r, info.URL, http.StatusFound)
	}))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	log.Printf("dmgr-web-backend listening on :%s (repo %s)", port, repo)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
