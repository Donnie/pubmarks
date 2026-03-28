# Pubmarks Dataset

A repository of daily updated datasets.

## Datasets

| Data | Source |
|------|--------|
| [S&P 500® with weights](./datasets/etfs/spy/latest.json) | [State Street® SPDR® S&P 500® ETF Trust](https://www.ssga.com/us/en/individual/etfs/state-street-spdr-sp-500-etf-trust-spy) |
| [S&P® Semiconductor with weights](./datasets/etfs/xsd/latest.json) | [State Street® SPDR® S&P® Semiconductor ETF](https://www.ssga.com/us/en/individual/etfs/state-street-spdr-sp-semiconductor-etf-xsd) |

## Expected schema

The shape below is best-effort documentation only. Source websites, HTML, and export formats change; keys can be missing, renamed, or parsed differently over time. Do not treat this as a stable API contract.

### Shared by `statestreet` and `blackrock`

Both tools print a single JSON object to stdout with:

| Key | Type | Notes |
|-----|------|--------|
| `date` | string | `YYYY-MM-DD` (holdings or fund-characteristics as-of, depending on the scraper). |
| `holdings` | array | Sorted by `name`. See per-tool rows below for line-item fields. |
| `holdings[].name` | string | Both tools. |
| `holdings[].ticker` | string | Both tools. Underlying security ticker (not the fund). |
| `holdings[].weight` | number | Both tools. Fund weight as a percentage of the portfolio (provider convention, not necessarily 0–1). |
| `holdings[].shares_held` | number | Both tools. Share count (or provider equivalent). |
| `base_currency` | string | Present when a primary price string can be parsed (ISO 4217 code, e.g. `USD`). |
| `price` (top-level) | number | Fund-level quote: parsed from `closing_price` (`statestreet`) or `nav` (`blackrock`). |
| `ticker` | string | Fund symbol (e.g. `SPY`, `SEMI`). Use it (lowercased) as `{ticker}` in `datasets/etfs/{ticker}/`. |

Additional top-level keys are merged in as strings unless noted. Key order in JSON is not guaranteed.
