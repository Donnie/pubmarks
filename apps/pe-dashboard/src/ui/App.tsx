import { useEffect, useMemo, useRef, useState } from "react";
import type { IChartApi, LineData, CandlestickData, WhitespaceData } from "lightweight-charts";
import {
  datasetBaseUrl,
  fetchManifest,
  fetchOhlcvYears,
  fetchPeYears,
  filterByStartDate,
  Manifest,
  PeriodKey,
  startDateForPeriod,
  toWeeklyOhlcv,
  yearsForPeriod
} from "../lib/datasets";
import { CandlesChart, LineChart } from "./ChartPane";

type LoadState =
  | { kind: "idle" }
  | { kind: "loading" }
  | { kind: "ready" }
  | { kind: "error"; message: string };

type CandleInterval = "1D" | "1W";

export function App() {
  const [loadState, setLoadState] = useState<LoadState>({ kind: "idle" });
  const [ticker, setTicker] = useState<string>("aapl");
  const [period, setPeriod] = useState<PeriodKey>("5Y");
  const [candleInterval, setCandleInterval] = useState<CandleInterval>("1W");

  const [tickers, setTickers] = useState<string[]>([]);
  const [manifest, setManifest] = useState<Manifest | null>(null);
  const [baseUrl, setBaseUrl] = useState<string | null>(null);

  const [candles, setCandles] = useState<CandlestickData<string>[]>([]);
  const [peLine, setPeLine] = useState<Array<LineData<string> | WhitespaceData<string>>>([]);

  const priceChartRef = useRef<IChartApi | null>(null);
  const peChartRef = useRef<IChartApi | null>(null);
  const syncingRef = useRef(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        setLoadState({ kind: "loading" });
        const m = await fetchManifest();
        const b = datasetBaseUrl(m);
        if (cancelled) return;
        setManifest(m);
        setBaseUrl(b);
        const t = Object.keys(m.datasets.stocks.tickers).sort();
        setTickers(t);
        if (!m.datasets.stocks.tickers[ticker]) {
          setTicker(t[0] ?? "aapl");
        }
        setLoadState({ kind: "ready" });
      } catch (e) {
        setLoadState({ kind: "error", message: e instanceof Error ? e.message : String(e) });
      }
    })();
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const subtitle = useMemo(() => {
    if (!manifest || !baseUrl) return "Loading manifest…";
    return `Data: ${baseUrl} • manifest: ${new Date(manifest.generatedAt).toLocaleString()}`;
  }, [manifest, baseUrl]);

  useEffect(() => {
    if (!manifest || !baseUrl) return;
    let cancelled = false;

    (async () => {
      try {
        setLoadState({ kind: "loading" });
        const stock = manifest.datasets.stocks.tickers[ticker];
        if (!stock) throw new Error(`Unknown ticker: ${ticker}`);

        const ohlcvYears = yearsForPeriod({
          minYear: stock.ohlcv.yearRange.min,
          maxYear: stock.ohlcv.yearRange.max,
          period
        });
        const peYears = yearsForPeriod({
          minYear: stock.peratio.yearRange.min,
          maxYear: stock.peratio.yearRange.max,
          period
        });

        const [ohlcv, pe] = await Promise.all([
          fetchOhlcvYears({ baseUrl, tickerLower: ticker, years: ohlcvYears }),
          fetchPeYears({ baseUrl, tickerLower: ticker, years: peYears })
        ]);
        if (cancelled) return;

        const endDate = ohlcv.length ? ohlcv[ohlcv.length - 1]!.time : pe.length ? pe[pe.length - 1]!.time : "1970-01-01";
        const startDate = startDateForPeriod({ endDate, period });

        const ohlcvFiltered = filterByStartDate(ohlcv, startDate);
        const ohlcvSeries = candleInterval === "1W" ? toWeeklyOhlcv(ohlcvFiltered) : ohlcvFiltered;
        const peFiltered = filterByStartDate(pe, startDate);

        setCandles(
          ohlcvSeries.map((b) => ({
            time: b.time,
            open: b.open,
            high: b.high,
            low: b.low,
            close: b.close
          }))
        );
        const peByTime = new Map<string, number>();
        for (const p of peFiltered) peByTime.set(p.time, p.pe);
        const candleTimes = new Set(ohlcvSeries.map((b) => b.time));

        const padded: Array<LineData<string> | WhitespaceData<string>> = ohlcvSeries.map((b) => {
          const v = peByTime.get(b.time);
          return v === undefined ? { time: b.time } : { time: b.time, value: v };
        });

        for (const p of peFiltered) {
          // If P/E timestamp doesn't align to weekly candle time (most won't), include it
          // so the actual quarterly points render.
          if (!candleTimes.has(p.time)) padded.push({ time: p.time, value: p.pe });
        }

        padded.sort((a, b) => a.time.localeCompare(b.time));
        setPeLine(padded);
        setLoadState({ kind: "ready" });
      } catch (e) {
        if (cancelled) return;
        setLoadState({ kind: "error", message: e instanceof Error ? e.message : String(e) });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [manifest, baseUrl, ticker, period, candleInterval]);

  useEffect(() => {
    const price = priceChartRef.current;
    const pe = peChartRef.current;
    if (!price || !pe) return;

    const priceTs = price.timeScale();
    const peTs = pe.timeScale();

    const onPriceRange = (range: Parameters<typeof priceTs.subscribeVisibleTimeRangeChange>[0] extends (r: infer R) => void ? R : never) => {
      if (!range || syncingRef.current) return;
      syncingRef.current = true;
      peTs.setVisibleRange(range);
      syncingRef.current = false;
    };

    const onPeRange = (range: Parameters<typeof peTs.subscribeVisibleTimeRangeChange>[0] extends (r: infer R) => void ? R : never) => {
      if (!range || syncingRef.current) return;
      syncingRef.current = true;
      priceTs.setVisibleRange(range);
      syncingRef.current = false;
    };

    priceTs.subscribeVisibleTimeRangeChange(onPriceRange);
    peTs.subscribeVisibleTimeRangeChange(onPeRange);

    return () => {
      priceTs.unsubscribeVisibleTimeRangeChange(onPriceRange);
      peTs.unsubscribeVisibleTimeRangeChange(onPeRange);
    };
  }, [candles.length, peLine.length]);

  return (
    <div className="container">
      <div className="header">
        <div className="title">
          <h1>Historical P/E dashboard</h1>
          <div className="subtitle">{subtitle}</div>
        </div>

        <div className="controls">
          <div className="control">
            <label>Ticker</label>
            <select value={ticker} onChange={(e) => setTicker(e.target.value)} disabled={loadState.kind === "loading"}>
              {tickers.map((t) => (
                <option key={t} value={t}>
                  {t.toUpperCase()}
                </option>
              ))}
            </select>
          </div>

          <div className="control">
            <label>Period</label>
            <select value={period} onChange={(e) => setPeriod(e.target.value as PeriodKey)} disabled={loadState.kind === "loading"}>
              <option value="1Y">1Y</option>
              <option value="5Y">5Y</option>
              <option value="10Y">10Y</option>
              <option value="MAX">MAX</option>
            </select>
          </div>

          <div className="control">
            <label>Candles</label>
            <select
              value={candleInterval}
              onChange={(e) => setCandleInterval(e.target.value as CandleInterval)}
              disabled={loadState.kind === "loading"}
            >
              <option value="1D">Daily</option>
              <option value="1W">Weekly</option>
            </select>
          </div>

          <div className="control" style={{ alignSelf: "end" }}>
            <button
              onClick={() => {
                priceChartRef.current?.timeScale().fitContent();
                peChartRef.current?.timeScale().fitContent();
              }}
              disabled={loadState.kind === "loading"}
            >
              Fit
            </button>
          </div>
        </div>
      </div>

      <div className="grid">
        <div className="panel">
          <div className="panelHeader">
            <div className="left">
              <div className="name">Price (OHLCV)</div>
              <div className="meta">{candles.length ? `${candles.length.toLocaleString()} bars` : "—"}</div>
            </div>
          </div>
          <div className="chartWrap">
            <CandlesChart
              height={420}
              data={candles}
              onChartReady={(c) => {
                priceChartRef.current = c;
              }}
            />
          </div>
        </div>

        <div className="panel">
          <div className="panelHeader">
            <div className="left">
              <div className="name">P/E (quarterly)</div>
              <div className="meta">{peLine.length ? `${peLine.length.toLocaleString()} points` : "—"}</div>
            </div>
          </div>
          <div className="chartWrap small">
            <LineChart
              height={220}
              data={peLine}
              color="#63b3ff"
              onChartReady={(c) => {
                peChartRef.current = c;
              }}
            />
          </div>
        </div>
      </div>

      {loadState.kind === "error" ? <div className="error">{loadState.message}</div> : null}
    </div>
  );
}
