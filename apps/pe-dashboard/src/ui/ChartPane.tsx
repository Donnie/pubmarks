import { useEffect, useMemo, useRef } from "react";
import {
  ColorType,
  createChart,
  IChartApi,
  LineData,
  CandlestickData,
  CandlestickSeries,
  LineSeries
} from "lightweight-charts";

type CommonProps = {
  height: number;
  onChartReady?: (chart: IChartApi) => void;
};

export function CandlesChart(props: CommonProps & { data: CandlestickData<string>[] }) {
  const elRef = useRef<HTMLDivElement | null>(null);
  const chartRef = useRef<IChartApi | null>(null);

  const options = useMemo(
    () => ({
      layout: {
        background: { type: ColorType.Solid as const, color: "transparent" },
        textColor: "rgba(230, 233, 242, 0.92)"
      },
      grid: {
        vertLines: { color: "rgba(255,255,255,0.06)" },
        horzLines: { color: "rgba(255,255,255,0.06)" }
      },
      timeScale: {
        borderColor: "rgba(255,255,255,0.12)"
      },
      rightPriceScale: {
        borderColor: "rgba(255,255,255,0.12)"
      },
      crosshair: {
        vertLine: { color: "rgba(99,179,255,0.35)" },
        horzLine: { color: "rgba(99,179,255,0.20)" }
      }
    }),
    []
  );

  useEffect(() => {
    if (!elRef.current) return;
    const chart = createChart(elRef.current, {
      ...options,
      width: elRef.current.clientWidth,
      height: props.height
    });
    chartRef.current = chart;

    const series = chart.addSeries(CandlestickSeries, {
      upColor: "#4ade80",
      downColor: "#fb7185",
      borderUpColor: "#4ade80",
      borderDownColor: "#fb7185",
      wickUpColor: "#4ade80",
      wickDownColor: "#fb7185"
    });
    series.setData(props.data);
    chart.timeScale().fitContent();
    props.onChartReady?.(chart);

    const ro = new ResizeObserver(() => {
      if (!elRef.current) return;
      chart.applyOptions({ width: elRef.current.clientWidth, height: props.height });
    });
    ro.observe(elRef.current);

    return () => {
      ro.disconnect();
      chart.remove();
      chartRef.current = null;
    };
  }, [options, props.data, props.height, props.onChartReady]);

  return <div ref={elRef} style={{ width: "100%", height: props.height }} />;
}

export function LineChart(
  props: CommonProps & { data: LineData<string>[]; color?: string; priceScaleId?: string }
) {
  const elRef = useRef<HTMLDivElement | null>(null);

  const options = useMemo(
    () => ({
      layout: {
        background: { type: ColorType.Solid as const, color: "transparent" },
        textColor: "rgba(230, 233, 242, 0.92)"
      },
      grid: {
        vertLines: { color: "rgba(255,255,255,0.06)" },
        horzLines: { color: "rgba(255,255,255,0.06)" }
      },
      timeScale: {
        borderColor: "rgba(255,255,255,0.12)"
      },
      rightPriceScale: {
        borderColor: "rgba(255,255,255,0.12)"
      },
      crosshair: {
        vertLine: { color: "rgba(99,179,255,0.35)" },
        horzLine: { color: "rgba(99,179,255,0.20)" }
      }
    }),
    []
  );

  useEffect(() => {
    if (!elRef.current) return;
    const chart = createChart(elRef.current, {
      ...options,
      width: elRef.current.clientWidth,
      height: props.height
    });

    const series = chart.addSeries(LineSeries, {
      color: props.color ?? "#63b3ff",
      lineWidth: 2,
      priceScaleId: props.priceScaleId
    });
    series.setData(props.data);
    chart.timeScale().fitContent();
    props.onChartReady?.(chart);

    const ro = new ResizeObserver(() => {
      if (!elRef.current) return;
      chart.applyOptions({ width: elRef.current.clientWidth, height: props.height });
    });
    ro.observe(elRef.current);

    return () => {
      ro.disconnect();
      chart.remove();
    };
  }, [options, props.color, props.data, props.height, props.onChartReady, props.priceScaleId]);

  return <div ref={elRef} style={{ width: "100%", height: props.height }} />;
}

