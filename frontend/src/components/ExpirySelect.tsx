import dayjs from "dayjs";
import { CalendarIcon } from "lucide-react";
import React from "react";
import { Calendar } from "src/components/ui/calendar";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "src/components/ui/popover";
import { cn } from "src/lib/utils";
import { expiryOptions } from "src/types";

const daysFromNow = (date?: Date) => {
  if (!date) {
    return undefined;
  }
  const now = dayjs();
  const targetDate = dayjs(date);
  return targetDate.diff(now, "day");
};

interface ExpiryProps {
  value?: Date | undefined;
  onChange: (expiryDate?: Date) => void;
}

const ExpirySelect: React.FC<ExpiryProps> = ({ value, onChange }) => {
  const [expiryDays, setExpiryDays] = React.useState(daysFromNow(value));
  const [customExpiry, setCustomExpiry] = React.useState(() => {
    const _daysFromNow = daysFromNow(value);
    return _daysFromNow !== undefined
      ? !Object.values(expiryOptions)
          .filter((value) => value !== 0)
          .includes(_daysFromNow)
      : false;
  });
  return (
    <>
      <p className="font-medium text-sm mb-2">Connection expiration</p>
      <div className="grid grid-cols-2 md:grid-cols-6 gap-2 text-xs">
        {Object.keys(expiryOptions).map((expiry) => {
          return (
            <button
              type="button"
              key={expiry}
              onClick={() => {
                setCustomExpiry(false);
                let date: Date | undefined;
                if (expiryOptions[expiry]) {
                  date = dayjs()
                    .add(expiryOptions[expiry], "day")
                    .endOf("day")
                    .toDate();
                }
                onChange(date);
                setExpiryDays(expiryOptions[expiry]);
              }}
              className={cn(
                "cursor-pointer rounded text-nowrap border-2 text-center p-4",
                !customExpiry && expiryDays == expiryOptions[expiry]
                  ? "border-primary"
                  : "border-muted"
              )}
            >
              {expiry}
            </button>
          );
        })}
        <Popover>
          <PopoverTrigger asChild>
            <button
              onClick={() => {}}
              className={cn(
                "flex items-center justify-center md:col-span-2 cursor-pointer rounded text-nowrap border-2 p-4",
                customExpiry ? "border-primary" : "border-muted"
              )}
            >
              <CalendarIcon className="mr-2 h-4 w-4" />
              <span className="truncate">
                {customExpiry && value
                  ? dayjs(value).format("DD MMMM YYYY")
                  : "Custom..."}
              </span>
            </button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0">
            <Calendar
              mode="single"
              disabled={{
                before: new Date(),
              }}
              selected={value}
              onSelect={(date?: Date) => {
                if (!date) {
                  return;
                }
                date.setHours(23, 59, 59);
                setCustomExpiry(true);
                onChange(date);
                setExpiryDays(daysFromNow(date));
              }}
              autoFocus
            />
          </PopoverContent>
        </Popover>
      </div>
    </>
  );
};

export default ExpirySelect;
