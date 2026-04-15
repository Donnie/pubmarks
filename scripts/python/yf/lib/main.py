"""CLI entrypoint: macrotrends-compatible `yf ohlcv`."""

from __future__ import annotations

import sys

import click

from lib.csvfmt import err, write_ohlcv_csv
from lib.ohlcv_fetch import fetch_daily_ohlcv
from lib.ticker_year import parse_ticker_year


def _need_subcommand() -> None:
    err(
        "specify a subcommand: ohlcv\n\n"
        "Examples:\n"
        "  yf ohlcv AAPL\n"
        "  yf ohlcv MSFT 2024\n"
        "  TICKER=AAPL yf ohlcv 2024\n\n"
        'Run "yf --help" for full usage.'
    )
    raise SystemExit(1)


@click.group(invoke_without_command=True)
@click.pass_context
def yf(ctx: click.Context) -> None:
    if ctx.invoked_subcommand is None:
        _need_subcommand()


@yf.command(
    "ohlcv",
    context_settings={"show_default": True},
    help="Download daily OHLCV as CSV to stdout (Yahoo Finance via yfinance).",
)
@click.argument("parts", nargs=-1)
def ohlcv_cmd(parts: tuple[str, ...]) -> None:
    """
    Ticker: TICKER environment variable and/or first argument.

    Year: YEAR environment and/or last argument; a single 4-digit argument sets year only (ticker from env).

    If year is omitted, all rows from the download are printed.
    """
    try:
        ticker, year, year_set = parse_ticker_year(parts)
    except ValueError as e:
        err(str(e))
        raise SystemExit(1) from e

    try:
        rows = fetch_daily_ohlcv(ticker, year=year if year_set else None, year_set=year_set)
    except Exception as e:
        err(str(e))
        raise SystemExit(1) from e

    write_ohlcv_csv(rows, sys.stdout)


def main() -> None:
    yf()


if __name__ == "__main__":
    main()
