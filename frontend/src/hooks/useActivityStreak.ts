import * as React from "react";

import { useTransactions } from "src/hooks/useTransactions";

export type WeekDay = {
  dateKey: string;
  weekdayLabel: string;
  hasActivity: boolean;
  isToday: boolean;
  isFuture: boolean;
};

export type UseActivityStreakResult = {
  isLoading: boolean;
  currentStreak: number;
  keptToday: boolean;
  weekDays: WeekDay[];
};

const TRANSACTION_LIMIT = 500;
const WEEKDAY_LABELS = ["Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"];

// TEMP STUB — seeds a randomized streak for visual preview.
// Revert before merging.
export const IS_STREAK_STUBBED = true;

function generateStubStreakLength(): number {
  const r = Math.random();
  if (r < 0.15) {
    return 0;
  }
  if (r < 0.25) {
    return 1;
  }
  if (r < 0.5) {
    return Math.floor(Math.random() * 10) + 2;
  }
  if (r < 0.8) {
    return Math.floor(Math.random() * 50) + 10;
  }
  return Math.floor(Math.random() * 500) + 50;
}

function formatLocalDateKey(date: Date, timeZone: string): string {
  return new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    timeZone,
  }).format(date);
}

function addDaysToKey(key: string, delta: number): string {
  const [year, month, day] = key.split("-").map(Number);
  const dt = new Date(year, month - 1, day);
  dt.setDate(dt.getDate() + delta);
  const y = dt.getFullYear();
  const m = String(dt.getMonth() + 1).padStart(2, "0");
  const d = String(dt.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

export function useActivityStreak(stubTick = 0): UseActivityStreakResult {
  const { data, isLoading } = useTransactions(
    undefined,
    false,
    TRANSACTION_LIMIT,
    1
  );

  return React.useMemo(() => {
    const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    const now = new Date();
    const todayKey = formatLocalDateKey(now, timeZone);

    const dayOfWeek = now.getDay(); // 0 = Sun … 6 = Sat
    const daysFromMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const mondayKey = addDaysToKey(todayKey, -daysFromMonday);

    const activeDays = new Set<string>();
    if (IS_STREAK_STUBBED) {
      const streakLen = generateStubStreakLength();
      const keptToday = streakLen > 0 && Math.random() > 0.3;
      const startOffset = keptToday ? 0 : 1;
      for (let i = 0; i < streakLen; i++) {
        activeDays.add(addDaysToKey(todayKey, -(i + startOffset)));
      }
    } else {
      for (const tx of data?.transactions ?? []) {
        if (tx.state !== "settled" || !tx.settledAt) {
          continue;
        }
        activeDays.add(formatLocalDateKey(new Date(tx.settledAt), timeZone));
      }
    }

    let currentStreak = 0;
    let cursor = todayKey;
    if (!activeDays.has(cursor)) {
      cursor = addDaysToKey(cursor, -1);
    }
    while (activeDays.has(cursor)) {
      currentStreak += 1;
      cursor = addDaysToKey(cursor, -1);
    }

    const weekDays: WeekDay[] = WEEKDAY_LABELS.map((label, i) => {
      const dateKey = addDaysToKey(mondayKey, i);
      return {
        dateKey,
        weekdayLabel: label,
        hasActivity: activeDays.has(dateKey),
        isToday: dateKey === todayKey,
        isFuture: dateKey > todayKey,
      };
    });

    return {
      isLoading: isLoading && !data && !IS_STREAK_STUBBED,
      currentStreak,
      keptToday: activeDays.has(todayKey),
      weekDays,
    };
    // stubTick intentionally re-seeds the memo to reroll the stub.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, isLoading, stubTick]);
}
