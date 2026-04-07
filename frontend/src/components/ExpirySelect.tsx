import dayjs from "dayjs";
import { CalendarIcon, XIcon } from "lucide-react";
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
  onClose?: () => void;
}

const ExpirySelect: React.FC<ExpiryProps> = ({ value, onChange, onClose }) => {
  const [expiryDays, setExpiryDays] = React.useState(daysFromNow(value));

  const isPreset =
    expiryDays !== undefined &&
    Object.values(expiryOptions)
      .filter((v) => v !== 0)
      .includes(expiryDays);

  return (
    <>
      <div className="flex items-center mb-2">
        <p className="font-medium text-sm">Connection expiration</p>
        {onClose && (
          <XIcon
            className="cursor-pointer w-4 ml-2 text-muted-foreground"
            onClick={onClose}
          />
        )}
      </div>
      <div className="grid grid-cols-3 gap-3 text-xs mb-3">
        {Object.keys(expiryOptions)
          .filter((expiry) => expiryOptions[expiry] !== 0)
          .map((expiry) => {
            return (
              <button
                type="button"
                key={expiry}
                onClick={() => {
                  const date = dayjs()
                    .add(expiryOptions[expiry], "day")
                    .endOf("day")
                    .toDate();
                  onChange(date);
                  setExpiryDays(expiryOptions[expiry]);
                }}
                className={cn(
                  "cursor-pointer rounded text-nowrap border-2 text-center p-3 py-4",
                  isPreset && expiryDays == expiryOptions[expiry]
                    ? "border-primary"
                    : "border-muted"
                )}
              >
                {expiry}
              </button>
            );
          })}
      </div>
      <div className="mb-4">
        <Popover>
          <PopoverTrigger asChild>
            <button
              type="button"
              className="flex items-center w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-left"
            >
              <span className={cn("flex-1", !value && "text-muted-foreground")}>
                {value ? dayjs(value).format("DD MMMM YYYY") : "Custom date"}
              </span>
              <CalendarIcon className="h-4 w-4 text-muted-foreground" />
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
