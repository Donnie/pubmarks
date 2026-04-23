#!/usr/bin/env python3
# /// script
# requires-python = ">=3.11"
# ///

"""
Build datasets/manifest.json by scanning the datasets/ tree.
Run from repository root: uv run scripts/python/manifest/main.py
"""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

# scripts/python/manifest/ -> repo root
ROOT = Path(__file__).resolve().parent.parent.parent.parent
DATASETS = ROOT / "datasets"
OUT = DATASETS / "manifest.json"

DATE_JSON = re.compile(r"^(\d{4}-\d{2}-\d{2})\.json$")
YEAR_DIR = re.compile(r"^\d{4}$")


def iso_now() -> str:
    return datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace(
        "+00:00", "Z"
    )


def scan_etfs(etfs_root: Path) -> dict:
    tickers: dict[str, dict] = {}
    if not etfs_root.is_dir():
        return {"root": "datasets/etfs", "tickers": tickers, "tickerCount": 0}

    for d in sorted(etfs_root.iterdir()):
        if not d.is_dir():
            continue
        sym = d.name.lower()
        dates: list[str] = []
        has_latest = False
        for f in d.iterdir():
            if not f.is_file():
                continue
            if f.name == "latest.json":
                has_latest = True
                continue
            m = DATE_JSON.match(f.name)
            if m:
                dates.append(m.group(1))
        dates.sort()
        tickers[sym] = {
            "hasLatestJson": has_latest,
            "datedSnapshotCount": len(dates),
            "datedSnapshots": dates,
        }
        if dates:
            tickers[sym]["snapshotDateRange"] = {
                "min": dates[0],
                "max": dates[-1],
            }
    return {
        "root": "datasets/etfs",
        "tickerCount": len(tickers),
        "conventions": {
            "layout": "One directory per ETF ticker; each holds dated snapshots and optional latest.json.",
            "files": "YYYY-MM-DD.json for daily snapshots; latest.json duplicates the most recent snapshot for stable URLs.",
        },
        "documentSchema": {
            "date": "string (ISO date of snapshot)",
            "holdings": "array of { name, ticker, weight, shares_held }",
        },
        "tickers": tickers,
    }


def scan_stocks(
    stocks_root: Path,
) -> tuple[dict, int, int]:
    """Returns (stocks section dict, ohlcv file count, peratio file count)."""
    tickers: dict[str, dict] = {}
    ohlcv_file_total = 0
    peratio_file_total = 0
    if stocks_root.is_dir():
        for d in sorted(stocks_root.iterdir()):
            if not d.is_dir():
                continue
            sym = d.name.lower()
            ohlcv_years: list[int] = []
            peratio_years: list[int] = []
            for ydir in d.iterdir():
                if not ydir.is_dir() or not YEAR_DIR.match(ydir.name):
                    continue
                year = int(ydir.name)
                if (ydir / "ohlcv.csv").is_file():
                    ohlcv_years.append(year)
                if (ydir / "peratio.csv").is_file():
                    peratio_years.append(year)
            ohlcv_years.sort()
            peratio_years.sort()
            entry: dict = {"ticker": sym}
            if ohlcv_years:
                ohlcv_file_total += len(ohlcv_years)
                entry["ohlcv"] = {
                    "file": "ohlcv.csv",
                    "yearRange": {"min": ohlcv_years[0], "max": ohlcv_years[-1]},
                }
            if peratio_years:
                peratio_file_total += len(peratio_years)
                entry["peratio"] = {
                    "file": "peratio.csv",
                    "yearRange": {"min": peratio_years[0], "max": peratio_years[-1]},
                }
            if ohlcv_years or peratio_years:
                tickers[sym] = entry

    return (
        {
            "root": "datasets/stocks",
            "tickerCount": len(tickers),
            "conventions": {
                "layout": "One directory per stock ticker; under each, one directory per calendar year (YYYY).",
                "files": "ohlcv.csv for daily OHLCV; peratio.csv for quarterly TTM P/E points.",
            },
            "series": {
                "ohlcv": {
                    "pathPattern": "datasets/stocks/{ticker}/{year}/ohlcv.csv",
                    "format": "csv",
                    "columns": ["date", "open", "high", "low", "close", "volume"],
                },
                "peratio": {
                    "pathPattern": "datasets/stocks/{ticker}/{year}/peratio.csv",
                    "format": "csv",
                    "columns": ["date", "stock_price", "ttm_net_eps", "pe_ratio"],
                },
            },
            "tickers": tickers,
        },
        ohlcv_file_total,
        peratio_file_total,
    )


def resolve_repository() -> str:
    r = os.environ.get("GITHUB_REPOSITORY", "").strip()
    if r:
        return r
    try:
        url = (
            subprocess.check_output(
                ["git", "remote", "get-url", "origin"],
                cwd=ROOT,
                text=True,
                stderr=subprocess.DEVNULL,
            )
            .strip()
            .rstrip("/")
        )
    except (OSError, subprocess.CalledProcessError):
        return "unknown/unknown"
    for prefix in ("git@github.com:", "https://github.com/", "ssh://git@github.com/"):
        if url.startswith(prefix):
            path = url[len(prefix) :].removesuffix(".git")
            return path
    return "unknown/unknown"


def build_manifest() -> dict:
    repo = resolve_repository()
    etfs = scan_etfs(DATASETS / "etfs")
    stocks, ohlcv_n, peratio_n = scan_stocks(DATASETS / "stocks")

    etf_files = 0
    for t in etfs.get("tickers", {}).values():
        etf_files += t.get("datedSnapshotCount", 0) + (1 if t.get("hasLatestJson") else 0)

    return {
        "schemaVersion": 1,
        "generatedAt": iso_now(),
        "repository": repo,
        "summary": {
            "etfs": {
                "tickers": etfs.get("tickerCount", 0),
                "jsonFiles": etf_files,
            },
            "stocks": {
                "tickers": stocks.get("tickerCount", 0),
                "ohlcvYearFiles": ohlcv_n,
                "peratioYearFiles": peratio_n,
            },
        },
        "access": {
            "jsdelivrCdnTemplate": f"https://cdn.jsdelivr.net/gh/{repo}@{{branch}}/{{path}}",
            "defaultBranch": "main",
        },
        "datasets": {
            "etfs": etfs,
            "stocks": stocks,
        },
    }


def main() -> int:
    if not DATASETS.is_dir():
        print("datasets/ not found; run from repository root", file=sys.stderr)
        return 1
    OUT.parent.mkdir(parents=True, exist_ok=True)
    manifest = build_manifest()
    with open(OUT, "w", encoding="utf-8") as f:
        json.dump(manifest, f, indent=2, ensure_ascii=False)
        f.write("\n")
    print(f"Wrote {OUT}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
