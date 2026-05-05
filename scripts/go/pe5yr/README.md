# pubmarks

Compute 5-year daily TTM valuation metrics for a stock ticker and print JSON.

## Usage

From this module directory (`apps/pubmarks`):

```bash
go run ./cmd/pubmarks AAPL
```

## Understanding the P/E values in the JSON output

This tool emits multiple “P/E-like” numbers because **plain P/E averaging breaks down** when earnings move near/through zero. The output includes **four different 5-year P/E lenses** (each answering a slightly different question).

- **`p_e_mean_5yr`** (alias: **`p_e_avg_5yr`**)
  - **Definition**: Arithmetic mean of **daily** `P/E` over the 5-year window, where daily `P/E = close / (TTM EPS as-of that day)`.
  - **Includes losses**: Yes (negative EPS days contribute negative P/E values).
  - **Caveat**: Can be dominated by days where EPS is very close to zero (P/E blows up to very large magnitude).

- **`p_e_earningsyield_5yr`** (the “E/P” best practice)
  - **Definition**: Compute daily earnings yield `E/P = EPS / close`, take the arithmetic mean over the window, then invert: `p_e_earningsyield_5yr = 1 / mean(E/P)`.
  - **Why it helps**: Price is rarely near zero, so `E/P` doesn’t explode the way `P/E` does when earnings approach 0.
  - **Interpretation**: Equivalent in spirit to a **harmonic-mean** style aggregation of P/E.
  - **Caveat**: If `mean(E/P) <= 0`, the inverted “P/E” is not economically meaningful (it indicates net losses over the window).

- **`p_e_shiller_5yr`** (normalized-earnings / “Shiller-style”)
  - **Definition**: Treat the 5-year period as one operating cycle by averaging earnings first:
    `p_e_shiller_5yr = price_last / mean(EPS over the window)`
    (uses the most recent close and the 5-year mean of the daily as-of EPS series; that mean EPS is computed internally and not emitted as a separate JSON field).
  - **Why it helps**: Smooths cyclical volatility by normalizing earnings rather than averaging ratios.
  - **Caveat**: If `mean(EPS over the window) <= 0`, the result is not meaningful as a valuation multiple.

- **`p_e_profitable_5yr`** (+ its companion **`p_e_lossy_5yr`**) (exclude + flag profitability)
  - **Definition**:
    - `p_e_profitable_5yr` is the arithmetic mean of **daily P/E only on days with EPS > 0**.
    - `p_e_lossy_5yr` is the arithmetic mean of **daily P/E only on days with EPS < 0**.
  - **Why it helps**: Lets you see “what valuation looked like when the business was profitable” without hiding that the company also had loss-making periods.
  - **Caveat**: These are conditional averages; interpret them alongside how often EPS was positive vs negative in the window.

## Notes on other P/E fields

- **`p_e_last`**: `price_last / eps_last` as of `end_date` (useful for “today’s” multiple, not a 5-year aggregation).
- **`p_e_median_5yr`** / **`p_e_mode_5yr`**: Robust/shape statistics of the daily P/E distribution (not substitutes for the four lenses above).
