"""OHLCV CSV format matching scripts/go/macrotrends (csvdata.go)."""

from __future__ import annotations

import csv
import io
import math
import sys
from typing import Iterable


OHLCV_HEADER = "date,open,high,low,close,volume"


def fmt_float(f: float) -> str:
    """Approximate Go strconv.FormatFloat(x, 'f', -1, 64) for OHLCV rows."""
    if not math.isfinite(f):
        raise ValueError(f"non-finite float: {f}")
    i = int(round(f))
    if abs(f - i) < 1e-9:
        return str(i)
    s = f"{f:.12f}".rstrip("0").rstrip(".")
    return s if s else "0"


def write_ohlcv_csv(rows: Iterable[tuple[str, float, float, float, float, float]], out: io.TextIOBase) -> None:
    w = csv.writer(out, lineterminator="\n")
    w.writerow(OHLCV_HEADER.split(","))
    for date, o, h, lo, c, vol in rows:
        w.writerow([date, fmt_float(o), fmt_float(h), fmt_float(lo), fmt_float(c), fmt_float(vol)])
    out.flush()


def err(msg: str) -> None:
    print(msg, file=sys.stderr)
