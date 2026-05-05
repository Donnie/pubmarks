# P/E dashboard (React)

Small React dashboard that charts:

- Price as **OHLCV candlesticks** (`datasets/stocks/{ticker}/{year}/ohlcv.csv`)
- **P/E ratio** as a simple **line chart** (`datasets/stocks/{ticker}/{year}/peratio.csv`)

## Run locally

```bash
cd apps/pe-dashboard
npm install
npm run dev
```

By default it fetches datasets from [GitHub Pages](https://donnie.github.io/pubmarks/) (the published site root is the repo’s `datasets/` folder, so CSV paths are `/stocks/...`, not `/datasets/stocks/...`). That works locally even though Vite does not serve `datasets/`.

## Optional: point at a different dataset host

Set `VITE_DATASET_BASE_URL` to the **site root** that contains `manifest.json` and `stocks/`, `etfs/` (same layout as `datasets/` in the repo).

Example:

```bash
VITE_DATASET_BASE_URL="https://donnie.github.io/pubmarks" npm run dev
```

## Build

```bash
npm run build
npm run preview
```

