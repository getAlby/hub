import { format } from "date-fns";
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
    return 0;
  }
  return Math.ceil(
    (new Date(date).getTime() - Date.now()) / (1000 * 60 * 60 * 24)
  );
};

interface ExpiryProps {
  value?: Date | undefined;
  onChange: (expiryDays: number) => void;
}

const ExpirySelect: React.FC<ExpiryProps> = ({ value, onChange }) => {
  const [expiryDays, setExpiryDays] = React.useState(daysFromNow(value));
  const [customExpiry, setCustomExpiry] = React.useState(
    daysFromNow(value)
      ? !Object.values(expiryOptions).includes(daysFromNow(value))
      : false
  );
  return (
    <div className="mb-6">
      <p className="font-medium text-sm mb-2">Connection expiration</p>
      <div className="grid grid-cols-2 md:grid-cols-6 gap-2 text-xs mb-4">
        {Object.keys(expiryOptions).map((expiry) => {
          return (
            <div
              key={expiry}
              onClick={() => {
                setCustomExpiry(false);
                onChange(expiryOptions[expiry]);
                setExpiryDays(expiryOptions[expiry]);
              }}
              className={cn(
                "cursor-pointer rounded text-nowrap border-2 text-center p-4 dark:text-white",
                !customExpiry && expiryDays == expiryOptions[expiry]
                  ? "border-primary"
                  : "border-muted"
              )}
            >
              {expiry}
            </div>
          );
        })}
        <Popover>
          <PopoverTrigger asChild>
            <div
              onClick={() => {}}
              className={cn(
                "flex items-center justify-center md:col-span-2 cursor-pointer rounded text-nowrap border-2 text-center px-3 py-4 dark:text-white",
                customExpiry ? "border-primary" : "border-muted"
              )}
            >
              <CalendarIcon className="mr-2 h-4 w-4" />
              <span className="truncate">
                {customExpiry && value ? format(value, "PPP") : "Custom..."}
              </span>
            </div>
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
                setCustomExpiry(true);
                onChange(daysFromNow(date));
                setExpiryDays(daysFromNow(date));
              }}
              initialFocus
            />
          </PopoverContent>
        </Popover>
      </div>
    </div>
  );
};

export default ExpirySelect;
