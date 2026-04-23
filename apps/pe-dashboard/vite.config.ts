import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  // This app is deployed under a subpath on GitHub Pages (and is also copied
  // under `datasets/pe-dashboard/` in this repo's Pages artifact). Using a
  // relative base keeps asset URLs working regardless of the hosting prefix.
  base: "./",
  plugins: [react()],
  server: {
    port: 5173
  }
});

