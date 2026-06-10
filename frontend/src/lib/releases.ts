// Shared release metadata. Used as a static fallback so the page works even when
// the Go backend isn't running; the client script upgrades to live data + counts
// when PUBLIC_API_BASE is reachable.

export const REPO = "Khinmmad/dmgr";
export const GITHUB_URL = "https://github.com/Khinmmad/dmgr";
export const RELEASES_PAGE = "https://github.com/Khinmmad/dmgr/releases/latest";
export const AUR_URL = "https://aur.archlinux.org/packages/dmgr-desktop";

export type Platform = "windows" | "macos" | "arch" | "debian" | "fedora" | "appimage";

export interface PlatformInfo {
  available: boolean;
  url?: string;
  filename?: string;
  command?: string; // for AUR
}

export const FALLBACK_VERSION = "v2.1.0";

const base = `${GITHUB_URL}/releases/download/${FALLBACK_VERSION}`;

export const FALLBACK: Record<Platform, PlatformInfo> = {
  windows: { available: false },
  macos: { available: false },
  arch: { available: true, url: AUR_URL, command: "paru -S dmgr-desktop" },
  debian: {
    available: true,
    url: `${base}/dmgr-desktop_2.1.0_amd64.deb`,
    filename: "dmgr-desktop_2.1.0_amd64.deb",
  },
  fedora: {
    available: true,
    url: `${base}/dmgr-desktop-2.1.0-1.x86_64.rpm`,
    filename: "dmgr-desktop-2.1.0-1.x86_64.rpm",
  },
  appimage: {
    available: true,
    url: `${base}/dmgr-desktop_2.1.0_amd64.AppImage`,
    filename: "dmgr-desktop_2.1.0_amd64.AppImage",
  },
};
