import { ArrowRightLeftIcon } from "lucide-react";
import * as React from "react";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "src/components/ui/field";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput,
  InputGroupText,
} from "src/components/ui/input-group";
import { Separator } from "src/components/ui/separator";
import { Skeleton } from "src/components/ui/skeleton";
import { BITCOIN_DISPLAY_FORMAT_BIP177 } from "src/constants";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";

type CurrencyInputMode = "bitcoin" | "fiat";

export type CurrencyInputContextRow = {
  label: string;
  amountSat?: number | null;
  value?: React.ReactNode;
};

type CurrencyInputFieldProps = Omit<
  React.ComponentProps<typeof InputGroupInput>,
  "max" | "min" | "onChange" | "step" | "type" | "value"
> & {
  contextRows?: CurrencyInputContextRow[];
  description?: React.ReactNode;
  error?: React.ReactNode;
  label?: React.ReactNode;
  maxSat?: number;
  minSat?: number;
  onValueSatChange: (valueSat: string) => void;
  valueSat: string;
};

const SATS_PER_BTC = 100_000_000;

function getNumericValue(value: string | number | null | undefined) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

function getCurrencyFractionDigits(currency: string) {
  try {
    return new Intl.NumberFormat("en-US", {
      currency,
      style: "currency",
    }).resolvedOptions().maximumFractionDigits;
  } catch {
    return 2;
  }
}

function getCurrencySymbol(currency: string) {
  try {
    return (
      new Intl.NumberFormat("en-US", {
        currency,
        style: "currency",
      })
        .formatToParts(0)
        .find((part) => part.type === "currency")?.value || currency
    );
  } catch {
    return currency;
  }
}

function formatFiatValue(
  amountSat: string | number | undefined,
  rate: number | undefined,
  currency: string | undefined
) {
  if (!rate || !currency) {
    return null;
  }

  return new Intl.NumberFormat("en-US", {
    currency,
    style: "currency",
  }).format((getNumericValue(amountSat) / SATS_PER_BTC) * rate);
}

function formatFiatInput(amountSat: string, rate: number, currency: string) {
  const fractionDigits = getCurrencyFractionDigits(currency);
  const amountFiat = (getNumericValue(amountSat) / SATS_PER_BTC) * rate;

  if (!amountFiat) {
    return "";
  }

  return amountFiat.toFixed(fractionDigits);
}

function formatBitcoinValue(
  amountSat: string | number | null | undefined,
  displayFormat: string | undefined
) {
  const formattedAmount = new Intl.NumberFormat().format(
    Math.floor(getNumericValue(amountSat))
  );

  if (displayFormat === BITCOIN_DISPLAY_FORMAT_BIP177) {
    return `₿${formattedAmount}`;
  }

  return `${formattedAmount} sats`;
}

export function CurrencyInputField({
  className,
  contextRows,
  description,
  disabled,
  error,
  id,
  label = "Amount",
  maxSat,
  minSat,
  onValueSatChange,
  required,
  valueSat,
  ...props
}: CurrencyInputFieldProps) {
  const generatedId = React.useId();
  const { data: info } = useInfo();
  const { data: bitcoinRate, error: bitcoinRateError } = useBitcoinRate(
    info?.currency
  );
  const [mode, setMode] = React.useState<CurrencyInputMode>("bitcoin");
  const [fiatValue, setFiatValue] = React.useState("");

  const currency = info?.currency || "USD";
  const rate = bitcoinRate?.rate_float;
  const canUseFiat = currency !== "SATS" && !!rate && !bitcoinRateError;
  const bitcoinUnit =
    info?.bitcoinDisplayFormat === BITCOIN_DISPLAY_FORMAT_BIP177 ? "₿" : "sats";
  const invalid =
    props["aria-invalid"] === true ||
    props["aria-invalid"] === "true" ||
    !!error;
  const inputId = id || generatedId;
  const isFiatMode = mode === "fiat";
  const inputValue = isFiatMode ? fiatValue : valueSat;
  const showLeadingUnit = isFiatMode || bitcoinUnit !== "sats";
  const alternateValue = isFiatMode
    ? formatBitcoinValue(valueSat, info?.bitcoinDisplayFormat)
    : formatFiatValue(valueSat, rate, currency);

  React.useEffect(() => {
    if (mode === "fiat" && !valueSat) {
      setFiatValue("");
    }
  }, [mode, valueSat]);

  function handleToggleMode() {
    if (!canUseFiat || disabled) {
      return;
    }

    if (mode === "bitcoin") {
      setFiatValue(formatFiatInput(valueSat, rate, currency));
      setMode("fiat");
      return;
    }

    setMode("bitcoin");
  }

  function handleChange(event: React.ChangeEvent<HTMLInputElement>) {
    const nextValue = event.target.value.trim();

    if (mode === "bitcoin") {
      onValueSatChange(nextValue);
      return;
    }

    setFiatValue(nextValue);

    if (!nextValue || !rate) {
      onValueSatChange("");
      return;
    }

    const amountFiat = Number(nextValue);
    if (!Number.isFinite(amountFiat)) {
      onValueSatChange("");
      return;
    }

    onValueSatChange(
      Math.max(0, Math.round((amountFiat / rate) * SATS_PER_BTC)).toString()
    );
  }

  function getModeBound(amountSat: number | undefined) {
    if (amountSat === undefined) {
      return undefined;
    }

    if (!isFiatMode || !rate) {
      return amountSat;
    }

    return (amountSat / SATS_PER_BTC) * rate;
  }

  return (
    <Field
      className={cn("w-full min-w-0", className)}
      data-disabled={disabled || undefined}
      data-invalid={invalid || undefined}
    >
      {label && <FieldLabel htmlFor={inputId}>{label}</FieldLabel>}
      <div
        className={cn(
          "w-full min-w-0 overflow-hidden rounded-md border border-input bg-background shadow-xs transition-[color,box-shadow]",
          "focus-within:border-ring focus-within:ring-[3px] focus-within:ring-ring/50",
          invalid &&
            "border-destructive ring-destructive/20 focus-within:border-destructive focus-within:ring-destructive/20"
        )}
      >
        <InputGroup className="h-9 min-w-0 rounded-none border-0 shadow-none">
          <InputGroupInput
            {...props}
            id={inputId}
            aria-invalid={invalid || undefined}
            className={cn(
              "sensitive slashed-zero min-w-0 [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none"
            )}
            disabled={disabled}
            inputMode="decimal"
            max={getModeBound(maxSat)}
            min={getModeBound(minSat)}
            onChange={handleChange}
            placeholder={isFiatMode ? "0.00" : "0"}
            required={required}
            step={isFiatMode ? "any" : 1}
            type="number"
            value={inputValue}
          />
          {showLeadingUnit && (
            <InputGroupAddon align="inline-start">
              <InputGroupText>
                {isFiatMode ? getCurrencySymbol(currency) : bitcoinUnit}
              </InputGroupText>
            </InputGroupAddon>
          )}
          <InputGroupAddon
            align="inline-end"
            className="!mr-0 min-w-0 self-stretch gap-0 py-0 pr-0"
          >
            <InputGroupText className="sensitive slashed-zero mr-2 min-w-0 max-w-28 truncate tabular-nums sm:max-w-none">
              {alternateValue ?? <Skeleton className="h-4 w-16" />}
            </InputGroupText>
            <InputGroupButton
              aria-label={
                isFiatMode ? "Enter amount in bitcoin" : "Enter amount in fiat"
              }
              disabled={!canUseFiat || disabled}
              onClick={handleToggleMode}
              size="icon-sm"
              className="!size-9 rounded-none border-l border-input p-0"
              title={
                isFiatMode ? "Enter amount in bitcoin" : "Enter amount in fiat"
              }
            >
              <ArrowRightLeftIcon className="size-4" />
            </InputGroupButton>
          </InputGroupAddon>
        </InputGroup>
        {!!contextRows?.length && (
          <>
            <Separator />
            <div className="flex min-w-0 flex-col gap-2 px-3 py-2 text-sm text-muted-foreground">
              {contextRows.map((row) => (
                <div
                  className="flex min-w-0 items-center justify-between gap-3"
                  key={row.label}
                >
                  <span className="truncate">{row.label}:</span>
                  <span className="sensitive slashed-zero min-w-0 max-w-[55%] truncate text-right tabular-nums">
                    {row.value ??
                      formatBitcoinValue(
                        row.amountSat,
                        info?.bitcoinDisplayFormat
                      )}
                  </span>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
      {description && <FieldDescription>{description}</FieldDescription>}
      {error && <FieldError>{error}</FieldError>}
    </Field>
  );
}
