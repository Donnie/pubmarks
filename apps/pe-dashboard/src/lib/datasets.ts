export type StockManifestTicker = {
  ticker: string;
  ohlcv: { file: string; yearRange: { min: number; max: number } };
  peratio: { file: string; yearRange: { min: number; max: number } };
};

export type Manifest = {
  schemaVersion: number;
  generatedAt: string;
  access: {
    jsdelivrCdnTemplate: string;
    defaultBranch: string;
  };
  datasets: {
    stocks: {
      tickers: Record<string, StockManifestTicker>;
    };
  };
};

export type OhlcvBar = {
  time: string; // YYYY-MM-DD
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
};

export type PePoint = {
  time: string; // YYYY-MM-DD
  pe: number;
  price: number; // stock_price in peratio series
};

function ensureOk(res: Response, url: string) {
  if (!res.ok) throw new Error(`Fetch failed ${res.status} for ${url}`);
}

export function datasetBaseUrl(manifest: Manifest): string {
  // Overrideable (useful if you later host the datasets elsewhere)
  const env = import.meta.env.VITE_DATASET_BASE_URL as string | undefined;
  if (env && env.trim().length > 0) return env.replace(/\/+$/, "");

  // Default to GitHub Pages (preferred over jsDelivr for this dashboard).
  // Pages root corresponds to the repo's `datasets/` folder contents, so:
  // - manifest: `${base}/manifest.json`
  // - data: `${base}/stocks/...`
  return "https://donnie.github.io/pubmarks";
}

export async function fetchManifest(base?: string): Promise<Manifest> {
  const env = (import.meta.env.VITE_DATASET_BASE_URL as string | undefined)?.trim();
  const candidates = [
    base ? `${base.replace(/\/+$/, "")}/manifest.json` : null,
    env ? `${env.replace(/\/+$/, "")}/manifest.json` : null,
    // When the dashboard is deployed under `/datasets/pe-dashboard/` (GitHub Pages artifact),
    // the manifest is at the Pages root: `/manifest.json`.
    "/manifest.json",
    // Also try relative to the dashboard path (`/pe-dashboard/` -> `../manifest.json`).
    "../manifest.json",
    "https://donnie.github.io/pubmarks/manifest.json"
  ].filter(Boolean) as string[];

  let lastErr: unknown = null;
  for (const url of candidates) {
    try {
      const res = await fetch(url, { cache: "no-cache" });
      ensureOk(res, url);
      return (await res.json()) as Manifest;
    } catch (e) {
      lastErr = e;
    }
  }

  throw lastErr instanceof Error ? lastErr : new Error(String(lastErr));
}

function parseCsvLines(text: string): string[][] {
  // Simple CSV parser (no quoted fields in these datasets).
  const lines = text.split(/\r?\n/).filter((l) => l.trim().length > 0);
  return lines.map((l) => l.split(","));
}

export async function fetchOhlcvYears(args: {
  baseUrl: string;
  tickerLower: string;
  years: number[];
}): Promise<OhlcvBar[]> {
  const all: OhlcvBar[] = [];

  for (const year of args.years) {
    const url = `${args.baseUrl}/stocks/${args.tickerLower}/${year}/ohlcv.csv`;
    const res = await fetch(url, { cache: "force-cache" });
    ensureOk(res, url);
    const text = await res.text();
    const rows = parseCsvLines(text);
    // header: date,open,high,low,close,volume
    for (let i = 1; i < rows.length; i++) {
      const [date, open, high, low, close, volume] = rows[i]!;
      if (!date) continue;
      all.push({
        time: date,
        open: Number(open),
        high: Number(high),
        low: Number(low),
        close: Number(close),
        volume: Number(volume)
      });
    }
  }

  all.sort((a, b) => a.time.localeCompare(b.time));
  return all;
}

export async function fetchPeYears(args: {
  baseUrl: string;
  tickerLower: string;
  years: number[];
}): Promise<PePoint[]> {
  const all: PePoint[] = [];

  for (const year of args.years) {
    const url = `${args.baseUrl}/stocks/${args.tickerLower}/${year}/peratio.csv`;
    const res = await fetch(url, { cache: "force-cache" });
    ensureOk(res, url);
    const text = await res.text();
    const rows = parseCsvLines(text);
    // header: date,stock_price,ttm_net_eps,pe_ratio
    for (let i = 1; i < rows.length; i++) {
      const [date, stockPrice, , peRatio] = rows[i]!;
      if (!date) continue;
      const pe = Number(peRatio);
      if (!Number.isFinite(pe)) continue;
      all.push({
        time: date,
        pe,
        price: Number(stockPrice)
      });
    }
  }

  all.sort((a, b) => a.time.localeCompare(b.time));
  return all;
}

export type PeriodKey = "1Y" | "5Y" | "10Y" | "MAX";

export function yearsForPeriod(args: {
  maxYear: number;
  minYear: number;
  period: PeriodKey;
}): number[] {
  let from = args.minYear;
  if (args.period === "1Y") from = Math.max(args.minYear, args.maxYear - 1);
  if (args.period === "5Y") from = Math.max(args.minYear, args.maxYear - 5);
  if (args.period === "10Y") from = Math.max(args.minYear, args.maxYear - 10);
  if (args.period === "MAX") from = args.minYear;

  const years: number[] = [];
  for (let y = from; y <= args.maxYear; y++) years.push(y);
  return years;
}

export function filterByStartDate<T extends { time: string }>(data: T[], startDate: string): T[] {
  const idx = data.findIndex((d) => d.time >= startDate);
  return idx === -1 ? [] : data.slice(idx);
}

export function startDateForPeriod(args: { endDate: string; period: PeriodKey }): string {
  if (args.period === "MAX") return "0000-01-01";
  const end = new Date(`${args.endDate}T00:00:00Z`);
  const start = new Date(end);
  const yearsBack = args.period === "1Y" ? 1 : args.period === "5Y" ? 5 : 10;
  start.setUTCFullYear(start.getUTCFullYear() - yearsBack);
  // keep ISO YYYY-MM-DD
  return start.toISOString().slice(0, 10);
}

