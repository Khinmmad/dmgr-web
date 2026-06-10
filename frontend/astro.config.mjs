import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";

// https://astro.build/config
// `base` / `site` are overridable via env so the same code deploys to a project
// subpath (GitHub Pages: /dmgr-web/) or to a root domain.
export default defineConfig({
  site: process.env.PUBLIC_SITE || "https://khinmmad.github.io",
  base: process.env.PUBLIC_BASE || "/",
  vite: {
    plugins: [tailwindcss()],
  },
});
