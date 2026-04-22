#!/usr/bin/env python3
# /// script
# requires-python = ">=3.12"
# dependencies = [
#   "playwright>=1.49",
# ]
# ///
"""
Export daily OHLCV data from a gettex.de instrument page.

The page loads history from the LSEG widget API; this script captures that
request in a browser context (to obtain the required `jwt` header), then
replays it with a custom date range and FIDs, writing CSV.

Run from repo root::

    uv run scripts/export_gettex_daily_csv.py -o apple_daily.csv

One-time Chromium install (if Playwright reports a missing browser; same
``playwright`` major as in the script block above)::

    uv run --with 'playwright>=1.49' python -m playwright install chromium
"""

from __future__ import annotations

import argparse
import csv
import datetime as dt
import sys

from playwright.sync_api import sync_playwright


def instrument_url(isin: str) -> str:
    isin = isin.strip().upper()
    return f"https://www.gettex.de/aktie/{isin}/"


def fetch_historical_rows(
    isin: str,
    timeout_ms: int,
) -> list[dict[str, str]]:
    url = instrument_url(isin)
    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        try:
            ctx = browser.new_context()
            page = ctx.new_page()

            captured_headers: dict[str, str] = {}

            def handle(route, request):
                h = request.headers
                if "jwt" in h:
                    captured_headers["jwt"] = h["jwt"]
                if "x-component-id" in h:
                    captured_headers["x-component-id"] = h["x-component-id"]
                if "referer" in h:
                    captured_headers["referer"] = h["referer"]
                if "origin" in h:
                    captured_headers["origin"] = h["origin"]
                route.continue_()

            page.route("**/rest/api/timeseries/historical**", handle)

            with page.expect_response(
                lambda r: "/rest/api/timeseries/historical" in r.url and r.status == 200,
                timeout=timeout_ms,
            ) as resp_info:
                page.goto(url, wait_until="domcontentloaded", timeout=timeout_ms)

            # Derive RIC from the request URL we observed during page load.
            observed_url = resp_info.value.url
            ric = observed_url.split("ric=", 1)[1].split("&", 1)[0]

            fids = ["_DATE_END", "OPEN_PRC", "HIGH_PRC", "LOW_PRC", "CLOSE_PRC", "ACVOL_1"]

            base = "https://lseg-widgets.financial.com/rest/api/timeseries/historical"
            # The endpoint appears to return whatever history it has available within the
            # requested window; using a very early fromDate requests the maximum period.
            to_date = dt.date.today()
            from_date = dt.date(1900, 1, 1)
            target = (
                f"{base}?ric={ric}"
                f"&fids={','.join(fids)}"
                "&samples=D&appendRecentData=all"
                f"&toDate={to_date.isoformat()}T23:59:59"
                f"&fromDate={from_date.isoformat()}T00:00:00"
            )

            res2 = page.request.get(target, headers=captured_headers, timeout=timeout_ms)
            payload = res2.json()
        finally:
            browser.close()

    data = payload.get("data")
    if not isinstance(data, list):
        raise RuntimeError(f"Unexpected payload shape: {type(data)}")
    return data


def write_csv(rows: list[dict[str, str]]) -> None:
    fieldnames = ("date", "open", "high", "low", "close", "volume")

    def row_to_out(row: dict[str, str]) -> dict[str, str]:
        return {
            "date": row.get("_DATE_END", ""),
            "open": row.get("OPEN_PRC", ""),
            "high": row.get("HIGH_PRC", ""),
            "low": row.get("LOW_PRC", ""),
            "close": row.get("CLOSE_PRC", ""),
            "volume": row.get("ACVOL_1", ""),
        }

    w = csv.DictWriter(sys.stdout, fieldnames=fieldnames, lineterminator="\n")
    w.writeheader()
    try:
        for row in rows:
            w.writerow(row_to_out(row))
    except BrokenPipeError:
        # Common when piping to `head`/`tail`.
        try:
            sys.stdout.close()
        finally:
            raise SystemExit(0)


def main() -> int:
    ap = argparse.ArgumentParser(
        description="Export gettex daily OHLCV (via LSEG widget API)."
    )
    ap.add_argument(
        "--isin",
        default="US0378331005",
        help="ISIN used in gettex URL (default: Apple)",
    )
    ap.add_argument(
        "--timeout-ms",
        type=int,
        default=60_000,
        help="Navigation / response wait timeout in ms",
    )
    args = ap.parse_args()

    rows = fetch_historical_rows(
        args.isin,
        args.timeout_ms,
    )
    write_csv(rows)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
