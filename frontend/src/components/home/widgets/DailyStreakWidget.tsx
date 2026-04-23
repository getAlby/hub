import { Share2Icon, ZapIcon } from "lucide-react";
import * as React from "react";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Skeleton } from "src/components/ui/skeleton";
import {
  IS_STREAK_STUBBED,
  useActivityStreak,
} from "src/hooks/useActivityStreak";
import { cn } from "src/lib/utils";

const MILESTONES = [7, 30, 100, 365, 1000];

type PillState = "kept" | "atRisk" | "start";

function getSubtitle(count: number): string {
  if (count === 0) {
    return "Send or receive a payment to start your streak.";
  }
  if (count === 1) {
    return "You've used lightning today — keep it going tomorrow.";
  }
  return `You've used lightning every day for ${count} days.`;
}

function getNextMilestone(
  streak: number
): { target: number; remaining: number } | null {
  for (const target of MILESTONES) {
    if (streak < target) {
      return { target, remaining: target - streak };
    }
  }
  return null;
}

function getPillState(currentStreak: number, keptToday: boolean): PillState {
  if (keptToday) {
    return "kept";
  }
  if (currentStreak > 0) {
    return "atRisk";
  }
  return "start";
}

const PILL_COPY: Record<PillState, string> = {
  kept: "Kept today",
  atRisk: "Keep it going",
  start: "Start today",
};

const PILL_CLASSES: Record<PillState, string> = {
  kept: "bg-green-500/15 text-green-700 dark:text-green-400",
  atRisk: "bg-amber-500/15 text-amber-700 dark:text-amber-400",
  start: "bg-muted text-muted-foreground",
};

const PILL_DOT_CLASSES: Record<PillState, string> = {
  kept: "bg-green-500",
  atRisk: "bg-amber-500",
  start: "bg-muted-foreground",
};

export function DailyStreakWidget() {
  const [stubTick, setStubTick] = React.useState(0);
  const { isLoading, currentStreak, keptToday, weekDays } =
    useActivityStreak(stubTick);

  const handleStubReroll = IS_STREAK_STUBBED
    ? () => setStubTick((t) => t + 1)
    : undefined;

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Daily streak</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-4">
            <Skeleton className="h-10 w-32" />
            <Skeleton className="h-4 w-64" />
            <div className="grid grid-cols-7 gap-2">
              {Array.from({ length: 7 }).map((_, i) => (
                <Skeleton key={i} className="h-12" />
              ))}
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  const pillState = getPillState(currentStreak, keptToday);
  const nextMilestone = getNextMilestone(currentStreak);

  return (
    <Card
      onClick={handleStubReroll}
      className={cn(handleStubReroll && "cursor-pointer select-none")}
      title={handleStubReroll ? "Click to randomize (stub)" : undefined}
    >
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle>Daily streak</CardTitle>
          <div
            className={cn(
              "inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium",
              PILL_CLASSES[pillState]
            )}
          >
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                PILL_DOT_CLASSES[pillState]
              )}
            />
            {PILL_COPY[pillState]}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-4">
          <div>
            <div className="flex items-baseline gap-2">
              <span className="text-4xl font-semibold tabular-nums">
                {currentStreak}
              </span>
              <span className="text-muted-foreground text-sm">
                {currentStreak === 1 ? "day" : "days"}
              </span>
            </div>
            <p className="text-muted-foreground mt-1 text-sm">
              {getSubtitle(currentStreak)}
            </p>
          </div>
          <div className="grid grid-cols-7 gap-2">
            {weekDays.map((day) => (
              <div
                key={day.dateKey}
                className="flex flex-col items-center gap-1.5"
              >
                <span className="text-muted-foreground text-xs">
                  {day.weekdayLabel}
                </span>
                <div
                  title={day.dateKey}
                  className={cn(
                    "flex h-10 w-full items-center justify-center rounded-md",
                    day.hasActivity
                      ? "bg-primary text-primary-foreground"
                      : "bg-muted text-muted-foreground",
                    day.isFuture && "opacity-40",
                    day.isToday &&
                      "ring-2 ring-primary ring-offset-2 ring-offset-background"
                  )}
                >
                  {day.hasActivity && (
                    <ZapIcon className="h-4 w-4 fill-current" />
                  )}
                </div>
              </div>
            ))}
          </div>
          <div className="flex items-end justify-between gap-2">
            {nextMilestone ? (
              <p className="text-muted-foreground text-xs">
                {nextMilestone.remaining}{" "}
                {nextMilestone.remaining === 1 ? "day" : "days"} to{" "}
                {nextMilestone.target}-day badge
              </p>
            ) : (
              <span />
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={(e) => e.stopPropagation()}
            >
              <Share2Icon className="mr-1.5 h-3.5 w-3.5" />
              Share
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
