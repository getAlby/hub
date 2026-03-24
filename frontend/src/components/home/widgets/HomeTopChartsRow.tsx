import React from "react";
import { Area, AreaChart, YAxis } from "recharts";
import { ChevronDownIcon, ChevronUpIcon, MinusIcon } from "lucide-react";
import useSWR from "swr";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from "src/components/ui/chart";
import { useBalances } from "src/hooks/useBalances";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";
import { ListTransactionsResponse, Transaction } from "src/types";
import { formatBitcoinAmount } from "src/utils/bitcoinFormatting";
import { request } from "src/utils/request";
import { cn } from "src/lib/utils";

const DAYS = 7;
const WINDOW_MS = DAYS * 24 * 60 * 60 * 1000;
const TX_PAGE_SIZE = 200;
const MAX_TX_PAGES = 40;

type Point = {
  day: string;
  value: number;
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

function getNetSatForTx(tx: Transaction) {
  const amount = satFromMsat(tx.amount);
  const fees = satFromMsat(tx.feesPaid);
  if (tx.type === "incoming") {
    return amount;
  }
  return -(amount + fees);
}

function getTxTimestamp(tx: Transaction) {
  return new Date(tx.settledAt || tx.updatedAt || tx.createdAt).getTime();
}

async function fetchSettledTransactionsForWindow(windowStart: number) {
  const results: Transaction[] = [];
  const seen = new Set<string>();
  let offset = 0;
  let totalCount = Number.POSITIVE_INFINITY;
  let page = 0;

  while (offset < totalCount && page < MAX_TX_PAGES) {
    const response = await request<ListTransactionsResponse>(
      `/api/transactions?limit=${TX_PAGE_SIZE}&offset=${offset}`
    );
    const transactions = response?.transactions || [];
    totalCount = response?.totalCount ?? totalCount;
    if (!transactions.length) {
      break;
    }

    let pageReachedOlderRange = false;
    for (const tx of transactions) {
      if (tx.state !== "settled") {
        continue;
      }
      const ts = getTxTimestamp(tx);
      if (ts < windowStart) {
        pageReachedOlderRange = true;
        continue;
      }
      const key = `${tx.paymentHash}:${tx.createdAt}:${tx.type}:${tx.amount}`;
      if (!seen.has(key)) {
        seen.add(key);
        results.push(tx);
      }
    }

    // The API returns newest-first; once page includes older-than-window settled txs
    // we have reached the lower boundary needed for accurate 7d charts.
    if (pageReachedOlderRange) {
      break;
    }

    offset += TX_PAGE_SIZE;
    page += 1;
  }

  return results.sort((a, b) => getTxTimestamp(a) - getTxTimestamp(b));
}

function flattenSeriesVisual(data: Point[], factor = 0.7) {
  if (data.length <= 1) {
    return data;
  }
  const avg = data.reduce((sum, point) => sum + point.value, 0) / data.length;
  return data.map((point) => ({
    ...point,
    value: avg + (point.value - avg) * factor,
  }));
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
    // Some APIs return unix microseconds.
    if (ts > 1e14) {
      ts = Math.floor(ts / 1000);
    }
    // Some APIs return unix seconds; convert to milliseconds for Date().
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

function buildPriceSeries(
  dailyRawPayloads: unknown[] | undefined,
  currencyCode: string,
  fallbackPrice?: number
) {
  const days = buildDaySeries();
  let lastKnown = fallbackPrice ?? 0;

  const series: Point[] = days.map((date, idx) => {
    const raw = dailyRawPayloads?.[idx];
    const points = parsePricePoints(raw, currencyCode);
    const dayEndTs = endOfDay(date).getTime();
    const pointsForDay = points.filter((point) => point.ts <= dayEndTs);
    const dayClose =
      pointsForDay.length > 0
        ? pointsForDay[pointsForDay.length - 1].price
        : points[points.length - 1]?.price;
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
  const stroke =
    changeValue >= 0 ? "var(--color-positive-foreground)" : "#f97316";
  const gradientId = `${title.toLowerCase().replace(/\s+/g, "-")}-area-gradient`;
  const IndicatorIcon =
    changeValue > 0
      ? ChevronUpIcon
      : changeValue < 0
        ? ChevronDownIcon
        : MinusIcon;
  const chartData = flattenSeriesVisual(data, 0.7);

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
        <ChartContainer
          className="h-[68px] w-full overflow-hidden"
          config={{
            value: {
              label: title,
              color: stroke,
            },
          }}
        >
          <AreaChart
            data={chartData}
            // Keep the chart hard-pinned to card edges.
            margin={{ left: -16, right: -16, top: 8, bottom: -10 }}
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
              padding={{ top: 0, bottom: 0 }}
            />
            <ChartTooltip
              content={<ChartTooltipContent />}
              cursor={false}
              isAnimationActive={false}
            />
            <Area
              type="monotone"
              dataKey="value"
              stroke={stroke}
              fill={`url(#${gradientId})`}
              fillOpacity={1}
              strokeWidth={2}
              isAnimationActive={false}
              baseLine={0}
            />
          </AreaChart>
        </ChartContainer>
      </CardContent>
    </Card>
  );
}

export function HomeTopChartsRow() {
  const { data: balances } = useBalances(true);
  const { data: info } = useInfo();
  const { data: bitcoinRate } = useBitcoinRate(true);
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
        historicalPriceEndpoints.map((endpoint) => request<unknown>(endpoint))
      );
      return responses;
    },
    { refreshInterval: 3000, refreshWhenHidden: true }
  );
  const [nowTs, setNowTs] = React.useState(() => Date.now());

  React.useEffect(() => {
    const id = window.setInterval(() => {
      setNowTs(Date.now());
    }, 3000);
    return () => window.clearInterval(id);
  }, []);

  const windowStart = nowTs - WINDOW_MS;
  const { data: settledTxsInWindow } = useSWR<Transaction[]>(
    info
      ? ["home-settled-transactions-window", Math.floor(windowStart / 60_000)]
      : null,
    () => fetchSettledTransactionsForWindow(windowStart),
    { refreshInterval: 5000, refreshWhenHidden: true }
  );

  if (!balances || !settledTxsInWindow || !info) {
    return null;
  }

  const totalBalanceSat = satFromMsat(
    msatFromSat(balances.onchain.total) + balances.lightning.totalSpendable
  );

  const hasIncomingDeposit = settledTxsInWindow.some(
    (tx) =>
      tx.state === "settled" &&
      tx.type === "incoming" &&
      satFromMsat(tx.amount) > 0
  );

  // Hide these cards during zero-balance onboarding.
  if (totalBalanceSat <= 0 && !hasIncomingDeposit) {
    return null;
  }

  const now = nowTs;
  const settledTxEntries = settledTxsInWindow.map((tx) => ({
    tx,
    timestamp: getTxTimestamp(tx),
  }));

  const netFlows7d = settledTxEntries.reduce(
    (sum, entry) => sum + getNetSatForTx(entry.tx),
    0
  );
  const startingBalance = totalBalanceSat - netFlows7d;

  const totalBalanceSeries: Point[] = (() => {
    let running = startingBalance;
    const series: Point[] = [
      {
        day: new Date(windowStart).toISOString().slice(11, 16),
        value: running,
      },
    ];
    for (const entry of settledTxEntries) {
      running += getNetSatForTx(entry.tx);
      series.push({
        day: new Date(entry.timestamp).toISOString().slice(11, 16),
        value: running,
      });
    }
    if (series.length === 1 || running !== totalBalanceSat) {
      series.push({
        day: new Date(now).toISOString().slice(11, 16),
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
    for (const entry of settledTxEntries) {
      running += getNetSatForTx(entry.tx);
      series.push({
        day: new Date(entry.timestamp).toISOString().slice(11, 16),
        value: running,
      });
    }
    if (series.length === 1 || running !== netFlows7d) {
      series.push({
        day: new Date(now).toISOString().slice(11, 16),
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
  const firstPricePoint = priceSeries.find((point) => point.value > 0);
  const lastPricePoint = [...priceSeries]
    .reverse()
    .find((point) => point.value > 0);
  const firstPrice = firstPricePoint?.value || bitcoinRate?.rate_float || 0;
  const lastPrice = lastPricePoint?.value || bitcoinRate?.rate_float || 0;
  const priceChangePct =
    firstPrice > 0 ? ((lastPrice - firstPrice) / firstPrice) * 100 : 0;

  const totalBalanceChangePct =
    startingBalance > 0
      ? ((totalBalanceSat - startingBalance) / startingBalance) * 100
      : 0;

  const currency = chartCurrency.toUpperCase();
  const displayFormat = info.bitcoinDisplayFormat || "sats";
  const currentPrice =
    priceSeries[priceSeries.length - 1]?.value || bitcoinRate?.rate_float || 0;
  const netFlowChangePct =
    totalBalanceSat > 0 ? (netFlows7d / totalBalanceSat) * 100 : 0;
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
        changeLabel={formatSignedPercent(priceChangePct)}
        changeValue={priceChangePct}
        data={priceSeries}
      />
    </div>
  );
}
