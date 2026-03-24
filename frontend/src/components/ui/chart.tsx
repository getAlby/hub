import * as React from "react";
import * as RechartsPrimitive from "recharts";
import { cn } from "src/lib/utils";

export type ChartConfig = Record<
  string,
  {
    label?: React.ReactNode;
    color?: string;
  }
>;

const ChartContext = React.createContext<{ config: ChartConfig } | null>(null);

function useChart() {
  const context = React.useContext(ChartContext);
  if (!context) {
    throw new Error("useChart must be used within a ChartContainer");
  }
  return context;
}

function ChartContainer({
  id,
  className,
  children,
  config,
  ...props
}: React.ComponentProps<"div"> & {
  config: ChartConfig;
  children: React.ComponentProps<
    typeof RechartsPrimitive.ResponsiveContainer
  >["children"];
}) {
  const uniqueId = React.useId();
  const chartId = `chart-${id || uniqueId.replace(/:/g, "")}`;

  return (
    <ChartContext.Provider value={{ config }}>
      <div
        data-slot="chart"
        data-chart={chartId}
        className={cn(
          "[&_.recharts-cartesian-grid_line[stroke='#ccc']]:stroke-border/70 [&_.recharts-curve.recharts-tooltip-cursor]:stroke-border [&_.recharts-dot[stroke='#fff']]:stroke-transparent [&_.recharts-layer]:outline-none",
          className
        )}
        {...props}
      >
        <style
          dangerouslySetInnerHTML={{
            __html: Object.entries(config)
              .map(([key, item]) =>
                item.color
                  ? `[data-chart=${chartId}]{--color-${key}:${item.color};}`
                  : ""
              )
              .join("\n"),
          }}
        />
        <RechartsPrimitive.ResponsiveContainer>
          {children}
        </RechartsPrimitive.ResponsiveContainer>
      </div>
    </ChartContext.Provider>
  );
}

function ChartTooltip({
  ...props
}: React.ComponentProps<typeof RechartsPrimitive.Tooltip>) {
  return <RechartsPrimitive.Tooltip {...props} />;
}

function ChartTooltipContent({
  active,
  payload,
  className,
  indicator = "dot",
}: React.ComponentProps<"div"> & {
  active?: boolean;
  payload?: Array<{
    dataKey?: string | number;
    color?: string;
    value?: unknown;
  }>;
  indicator?: "dot" | "line";
}) {
  const { config } = useChart();

  if (!active || !payload?.length) {
    return null;
  }

  return (
    <div
      className={cn(
        "grid min-w-[8rem] gap-1 rounded-md border bg-popover px-2.5 py-1.5 text-xs shadow-md",
        className
      )}
    >
      {payload.map((item, idx) => {
        const key = String(item.dataKey || `value-${idx}`);
        const itemConfig = config[key];
        const color = item.color || `var(--color-${key})`;
        return (
          <div key={key} className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-1.5 text-muted-foreground">
              {indicator === "dot" ? (
                <span
                  className="size-2 shrink-0 rounded-[2px]"
                  style={{ backgroundColor: color }}
                />
              ) : (
                <span
                  className="h-0.5 w-3 shrink-0"
                  style={{ backgroundColor: color }}
                />
              )}
              {itemConfig?.label || key}
            </div>
            <span className="font-medium text-foreground">
              {typeof item.value === "number"
                ? item.value.toLocaleString(undefined, {
                    maximumFractionDigits: 2,
                  })
                : String(item.value ?? "")}
            </span>
          </div>
        );
      })}
    </div>
  );
}

export { ChartContainer, ChartTooltip, ChartTooltipContent };
