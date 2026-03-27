import { ChevronDownIcon, ChevronUpIcon, MinusIcon } from "lucide-react";
import React from "react";
import { Area, AreaChart, ResponsiveContainer, YAxis } from "recharts";
import useSWR from "swr";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useBalances } from "src/hooks/useBalances";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";
import { formatBitcoinAmount } from "src/utils/bitcoinFormatting";
import { request } from "src/utils/request";

const DAYS = 7;
const WINDOW_MS = DAYS * 24 * 60 * 60 * 1000;
const CHART_REFRESH_MS = 5_000;

type Point = {
  day: string;
  value: number;
};

type HomeChartsResponse = {
  txPoints: Array<{
    timestamp: number;
    netSat: number;
  }>;
  hasIncomingDeposit: boolean;
  netFlowsSat: number;
};

function startOfDay(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

function endOfDay(date: Date) {
  return new Date(
    date.getFullYear(),
    date.getMonth(),
    date.getDate(),
    23,
    59,
    59,
    999
  );
}

function dayKey(date: Date) {
  return startOfDay(date).toISOString().slice(0, 10);
}

function buildDaySeries() {
  const today = startOfDay(new Date());
  const dates: Date[] = [];
  for (let i = DAYS - 1; i >= 0; i--) {
    const d = new Date(today);
    d.setDate(today.getDate() - i);
    dates.push(d);
  }
  return dates;
}

function satFromMsat(msat = 0) {
  return msat / 1000;
}

function msatFromSat(sat = 0) {
  return sat * 1000;
}

function formatBitcoinBySettings(
  valueSat: number,
  displayFormat: "sats" | "bip177"
) {
  return formatBitcoinAmount(Math.round(valueSat) * 1000, displayFormat, true);
}

function formatSignedPercent(value: number) {
  const sign = value > 0 ? "+" : value < 0 ? "-" : "";
  return `${sign}${Math.abs(value).toFixed(2)}%`;
}

function getChangeClass(delta: number) {
  if (delta > 0) {
    return "text-positive-foreground";
  }
  if (delta < 0) {
    return "text-orange-500";
  }
  return "text-muted-foreground";
}

function parsePricePoints(input: unknown, currencyCode: string) {
  const points: Array<{ ts: number; price: number }> = [];
  const upper = currencyCode.toUpperCase();
  const lower = currencyCode.toLowerCase();
  const normalizeTimestamp = (raw: unknown) => {
    let ts =
      typeof raw === "number"
        ? raw
        : typeof raw === "string"
          ? Number(raw)
          : NaN;
    if (!Number.isFinite(ts)) {
      return NaN;
    }
    if (ts > 1e14) {
      ts = Math.floor(ts / 1000);
    }
    if (ts > 0 && ts < 1e12) {
      ts *= 1000;
    }
    return ts;
  };

  const visit = (value: unknown) => {
    if (!value) {
      return;
    }
    if (Array.isArray(value)) {
      if (value.length >= 2) {
        const maybeTs = normalizeTimestamp(value[0]);
        const maybePrice =
          typeof value[1] === "number"
            ? value[1]
            : typeof value[1] === "string"
              ? Number(value[1])
              : NaN;
        if (
          Number.isFinite(maybeTs) &&
          Number.isFinite(maybePrice) &&
          maybePrice > 0
        ) {
          points.push({ ts: maybeTs, price: maybePrice });
          return;
        }
      }
      value.forEach(visit);
      return;
    }
    if (typeof value !== "object") {
      return;
    }
    const item = value as Record<string, unknown>;
    const rawTs =
      item.time ?? item.timestamp ?? item.date ?? item.t ?? item.createdAt;
    const ts = normalizeTimestamp(rawTs);
    const parsedDateTs = typeof rawTs === "string" ? Date.parse(rawTs) : NaN;
    const finalTs = Number.isFinite(ts) ? ts : parsedDateTs;

    const rawPrice =
      item[upper] ?? item[lower] ?? item.price ?? item.usd ?? item.USD;
    const price =
      typeof rawPrice === "number"
        ? rawPrice
        : typeof rawPrice === "string"
          ? Number(rawPrice)
          : NaN;

    if (Number.isFinite(finalTs) && Number.isFinite(price) && price > 0) {
      points.push({ ts: finalTs, price });
    }
    Object.values(item).forEach(visit);
  };

  visit(input);
  points.sort((a, b) => a.ts - b.ts);
  return points;
}

function getDayClosePrice(
  points: Array<{ ts: number; price: number }>,
  dayEndTs: number
) {
  const pointsForDay = points.filter((point) => point.ts <= dayEndTs);
  const dayClose =
    pointsForDay.length > 0
      ? pointsForDay[pointsForDay.length - 1].price
      : points[points.length - 1]?.price;
  return dayClose && dayClose > 0 ? dayClose : undefined;
}

function buildPriceSeries(
  dailyRawPayloads: unknown[] | undefined,
  currencyCode: string,
  fallbackPrice?: number
) {
  const days = buildDaySeries();
  let lastKnown = fallbackPrice ?? 0;

  const series: Point[] = days.map((date, idx) => {
    const points = parsePricePoints(dailyRawPayloads?.[idx], currencyCode);
    const dayClose = getDayClosePrice(points, endOfDay(date).getTime());
    if (dayClose && dayClose > 0) {
      lastKnown = dayClose;
    }

    return {
      day: dayKey(date).slice(5),
      value: lastKnown,
    };
  });

  if (!series.some((point) => point.value > 0) && fallbackPrice) {
    return days.map((date) => ({
      day: dayKey(date).slice(5),
      value: fallbackPrice,
    }));
  }

  return series;
}

function MetricCard({
  title,
  value,
  changeLabel,
  changeValue,
  data,
}: {
  title: string;
  value: string;
  changeLabel: string;
  changeValue: number;
  data: Point[];
}) {
  const valueColor = getChangeClass(changeValue);
  const stroke = changeValue >= 0 ? "#10b981" : "#f97316";
  const gradientId = `${title.toLowerCase().replace(/\s+/g, "-")}-sparkline-gradient`;
  const IndicatorIcon =
    changeValue > 0
      ? ChevronUpIcon
      : changeValue < 0
        ? ChevronDownIcon
        : MinusIcon;

  return (
    <Card className="gap-0 overflow-hidden rounded-[14px] py-0 shadow-none">
      <CardHeader className="px-5 pb-0 pt-4">
        <CardTitle className="text-base font-semibold">{title}</CardTitle>
      </CardHeader>
      <CardContent className="px-0 pb-0 pt-2">
        <div className="mb-1 flex items-end justify-between gap-3 px-5">
          <p className="text-2xl leading-none font-medium tracking-tight text-foreground">
            {value}
          </p>
          <CardDescription
            className={cn(
              "mb-1 flex items-center gap-1 text-xs font-medium",
              valueColor
            )}
          >
            <IndicatorIcon className="size-3.5 shrink-0" />
            <span>{changeLabel}</span>
            <span className="text-muted-foreground">7d</span>
          </CardDescription>
        </div>
        <div className="h-[68px] w-full overflow-hidden pt-4">
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart
              data={data}
              margin={{ top: 0, right: -10, left: -10, bottom: -10 }}
            >
              <defs>
                <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={stroke} stopOpacity={0.4} />
                  <stop offset="65%" stopColor={stroke} stopOpacity={0.14} />
                  <stop offset="100%" stopColor={stroke} stopOpacity={0.03} />
                </linearGradient>
              </defs>
              <YAxis
                hide
                domain={["dataMin", "dataMax"]}
                padding={{ top: 6 }}
              />
              <Area
                type="monotone"
                dataKey="value"
                stroke={stroke}
                fill={`url(#${gradientId})`}
                fillOpacity={1}
                strokeWidth={2}
                isAnimationActive={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}

export function HomeTopChartsRow() {
  const { data: balances } = useBalances();
  const { data: info } = useInfo();
  const { data: bitcoinRate } = useBitcoinRate();
  const priceHistoryCurrency =
    (info?.currency || "USD").toUpperCase() === "SATS"
      ? "USD"
      : info?.currency || "USD";
  const daySnapshots = buildDaySeries();
  const historicalPriceEndpoints = daySnapshots.map((day) => {
    const ts = Math.floor(endOfDay(day).getTime() / 1000);
    const mempoolEndpoint = `/v1/historical-price?currency=${encodeURIComponent(
      priceHistoryCurrency
    )}&timestamp=${ts}`;
    return `/api/mempool?endpoint=${encodeURIComponent(mempoolEndpoint)}`;
  });
  const { data: dailyPricePayloads } = useSWR<unknown[]>(
    info ? ["home-bitcoin-price-daily", ...historicalPriceEndpoints] : null,
    async () => {
      const responses = await Promise.all(
        historicalPriceEndpoints.map(async (endpoint) => {
          try {
            const response = await request<unknown>(endpoint);
            return response ?? {};
          } catch {
            return {};
          }
        })
      );
      return responses;
    },
    { refreshInterval: CHART_REFRESH_MS, refreshWhenHidden: true }
  );
  const [nowTs, setNowTs] = React.useState<number | null>(null);

  React.useEffect(() => {
    setNowTs(Date.now());
  }, []);
  const windowStart = (nowTs ?? 0) - WINDOW_MS;
  const fromSeconds = Math.floor(windowStart / 1000);
  const { data: homeChartsData } = useSWR<HomeChartsResponse>(
    info && nowTs !== null ? ["home-charts-data", fromSeconds] : null,
    async () => {
      const response = await request<HomeChartsResponse>(
        `/api/home/charts?from=${fromSeconds}`
      );
      if (!response) {
        throw new Error("Missing home chart payload");
      }
      return response;
    },
    { refreshInterval: CHART_REFRESH_MS, refreshWhenHidden: true }
  );

  if (!balances || !homeChartsData || !info || nowTs === null) {
    return null;
  }

  const totalBalanceSat = satFromMsat(
    msatFromSat(balances.onchain.total) + balances.lightning.totalSpendable
  );

  if (totalBalanceSat <= 0 && !homeChartsData.hasIncomingDeposit) {
    return null;
  }

  const txPoints = homeChartsData.txPoints || [];
  const netFlows7d = homeChartsData.netFlowsSat || 0;
  const startingBalance = totalBalanceSat - netFlows7d;

  const totalBalanceSeries: Point[] = (() => {
    let running = startingBalance;
    const series: Point[] = [
      {
        day: new Date(windowStart).toISOString().slice(11, 16),
        value: running,
      },
    ];
    for (const point of txPoints) {
      running += point.netSat;
      series.push({
        day: new Date(point.timestamp).toISOString().slice(11, 16),
        value: running,
      });
    }
    if (series.length === 1 || running !== totalBalanceSat) {
      series.push({
        day: new Date(nowTs).toISOString().slice(11, 16),
        value: totalBalanceSat,
      });
    }
    return series;
  })();

  const netFlowSeries: Point[] = (() => {
    let running = 0;
    const series: Point[] = [
      { day: new Date(windowStart).toISOString().slice(11, 16), value: 0 },
    ];
    for (const point of txPoints) {
      running += point.netSat;
      series.push({
        day: new Date(point.timestamp).toISOString().slice(11, 16),
        value: running,
      });
    }
    if (series.length === 1 || running !== netFlows7d) {
      series.push({
        day: new Date(nowTs).toISOString().slice(11, 16),
        value: netFlows7d,
      });
    }
    return series;
  })();

  const chartCurrency = priceHistoryCurrency;
  const priceSeries = buildPriceSeries(
    dailyPricePayloads,
    chartCurrency,
    bitcoinRate?.rate_float
  );
  const dayCloses = daySnapshots
    .map((day, idx) => {
      const points = parsePricePoints(dailyPricePayloads?.[idx], chartCurrency);
      return getDayClosePrice(points, endOfDay(day).getTime());
    })
    .filter((value): value is number => typeof value === "number");
  const firstPrice = dayCloses[0] || bitcoinRate?.rate_float || 0;
  const lastPrice =
    dayCloses[dayCloses.length - 1] || bitcoinRate?.rate_float || 0;
  const hasPriceChangeData =
    dayCloses.length >= 2 && firstPrice > 0 && lastPrice > 0;
  const priceChangePct = hasPriceChangeData
    ? ((lastPrice - firstPrice) / firstPrice) * 100
    : 0;

  const currency = chartCurrency.toUpperCase();
  const displayFormat = info.bitcoinDisplayFormat || "sats";
  const currentPrice =
    priceSeries[priceSeries.length - 1]?.value || bitcoinRate?.rate_float || 0;
  const netFlowChangePct =
    totalBalanceSat > 0 ? (netFlows7d / totalBalanceSat) * 100 : 0;
  const totalBalanceChangePct =
    startingBalance > 0
      ? ((totalBalanceSat - startingBalance) / startingBalance) * 100
      : netFlowChangePct;
  const priceLabel = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency === "SATS" ? "USD" : currency,
    maximumFractionDigits: 2,
  }).format(currentPrice);

  return (
    <div className="grid grid-cols-1 gap-3 xl:grid-cols-3">
      <MetricCard
        title="Total Balance"
        value={formatBitcoinBySettings(totalBalanceSat, displayFormat)}
        changeLabel={formatSignedPercent(totalBalanceChangePct)}
        changeValue={totalBalanceSat - startingBalance}
        data={totalBalanceSeries}
      />
      <MetricCard
        title="Net Flows"
        value={formatBitcoinBySettings(Math.abs(netFlows7d), displayFormat)}
        changeLabel={formatSignedPercent(netFlowChangePct)}
        changeValue={netFlowChangePct}
        data={netFlowSeries}
      />
      <MetricCard
        title="Bitcoin Price"
        value={priceLabel}
        changeLabel={
          hasPriceChangeData ? formatSignedPercent(priceChangePct) : "n/a"
        }
        changeValue={priceChangePct}
        data={priceSeries}
      />
    </div>
  );
}
