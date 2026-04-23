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

By default it fetches the datasets from jsDelivr (so it works even though `datasets/` is not served by Vite).

## Optional: point at a different dataset host

Set `VITE_DATASET_BASE_URL` to the host root that contains `/datasets/...`.

Examples:

```bash
VITE_DATASET_BASE_URL="https://cdn.jsdelivr.net/gh/Donnie/pubmarks@main" npm run dev
```

## Build

```bash
npm run build
npm run preview
```

