#!/usr/bin/env python3
"""Split a stock CSV (date column) into datasets/stocks/<ticker>/<year>/<out_basename>."""

import csv
import os
import sys


def main() -> None:
    if len(sys.argv) != 4:
        print("usage: split_stock_csv_by_year.py <ticker_lower> <input_csv> <out_basename>", file=sys.stderr)
        sys.exit(2)
    ticker_lower = sys.argv[1]
    path = sys.argv[2]
    out_basename = sys.argv[3]
    root = os.path.join("datasets", "stocks", ticker_lower)
    rows_by_year: dict[str, list[dict[str, str]]] = {}
    with open(path, newline="", encoding="utf-8") as f:
        r = csv.DictReader(f)
        fieldnames = r.fieldnames
        if not fieldnames:
            print("empty or headerless CSV", file=sys.stderr)
            sys.exit(1)
        fn = [x.strip() for x in fieldnames]
        date_key = "date" if "date" in fn else (fn[0] if fn else None)
        if not date_key:
            print("could not resolve date column", file=sys.stderr)
            sys.exit(1)
        for row in r:
            d = (row.get(date_key) or "").strip()
            if len(d) < 4:
                continue
            y = d[:4]
            if not y.isdigit():
                continue
            rows_by_year.setdefault(y, []).append(row)
    if not rows_by_year:
        print("no data rows after parse", file=sys.stderr)
        sys.exit(1)
    for y in sorted(rows_by_year.keys()):
        ydir = os.path.join(root, y)
        os.makedirs(ydir, exist_ok=True)
        out_path = os.path.join(ydir, out_basename)
        with open(out_path, "w", newline="", encoding="utf-8") as f:
            w = csv.DictWriter(f, fieldnames=fieldnames, lineterminator="\n")
            w.writeheader()
            w.writerows(rows_by_year[y])
        print("wrote", out_path, f"({len(rows_by_year[y])} rows)")


if __name__ == "__main__":
    main()
