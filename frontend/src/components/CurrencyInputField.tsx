import * as React from "react";
import { toast } from "sonner";
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
} from "src/components/ui/input-group";
import { Skeleton } from "src/components/ui/skeleton";
import { BITCOIN_DISPLAY_FORMAT_BIP177 } from "src/constants";
import { useBitcoinRate } from "src/hooks/useBitcoinRate";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";

type CurrencyInputMode = "bitcoin" | "fiat";
type BitcoinDenomination = "sats" | "btc";

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
  displayFormat: string | undefined,
  denomination: BitcoinDenomination = "sats"
) {
  const { amount, unit } = formatBitcoinValueParts(
    amountSat,
    displayFormat,
    denomination
  );

  if (unit === "₿") {
    return `${unit}${amount}`;
  }

  return `${amount} ${unit}`;
}

function formatBitcoinValueParts(
  amountSat: string | number | null | undefined,
  displayFormat: string | undefined,
  denomination: BitcoinDenomination = "sats"
) {
  if (denomination === "btc") {
    return {
      amount: formatBtcDisplay(amountSat),
      unit: "BTC",
    };
  }

  const formattedAmount = new Intl.NumberFormat().format(
    Math.floor(getNumericValue(amountSat))
  );

  if (displayFormat === BITCOIN_DISPLAY_FORMAT_BIP177) {
    return {
      amount: formattedAmount,
      unit: "₿",
    };
  }

  return {
    amount: formattedAmount,
    unit: "sats",
  };
}

function formatBtcDisplay(amountSat: string | number | null | undefined) {
  return (getNumericValue(amountSat) / SATS_PER_BTC).toFixed(8);
}

function formatBtcInput(amountSat: string | number | null | undefined) {
  const amount = getNumericValue(amountSat);

  if (!amount) {
    return "";
  }

  return (amount / SATS_PER_BTC).toFixed(8);
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
  const [bitcoinDenomination, setBitcoinDenomination] =
    React.useState<BitcoinDenomination>("sats");
  const [btcValue, setBtcValue] = React.useState("");

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
  const isBtcDenominated = bitcoinDenomination === "btc";
  const inputValue = isFiatMode
    ? fiatValue
    : isBtcDenominated
      ? btcValue
      : valueSat;
  const alternateBitcoinValue = formatBitcoinValueParts(
    valueSat,
    info?.bitcoinDisplayFormat,
    bitcoinDenomination
  );
  const alternateValue = isFiatMode
    ? formatBitcoinValue(
        valueSat,
        info?.bitcoinDisplayFormat,
        bitcoinDenomination
      )
    : formatFiatValue(valueSat, rate, currency);

  React.useEffect(() => {
    if (mode === "fiat" && !valueSat) {
      setFiatValue("");
    }
  }, [mode, valueSat]);

  React.useEffect(() => {
    if (mode === "bitcoin" && isBtcDenominated && !valueSat) {
      setBtcValue("");
    }
  }, [isBtcDenominated, mode, valueSat]);

  function handleToggleMode() {
    if (disabled) {
      return;
    }

    if (mode === "bitcoin") {
      if (!canUseFiat) {
        return;
      }

      setFiatValue(formatFiatInput(valueSat, rate, currency));
      setMode("fiat");
      return;
    }

    if (isBtcDenominated) {
      setBtcValue(formatBtcInput(valueSat));
    }

    setMode("bitcoin");
  }

  function handleAlternateValueClick() {
    if (disabled || isFiatMode || !canUseFiat) {
      return;
    }

    handleToggleMode();
  }

  function handleToggleBitcoinDenomination() {
    if (disabled) {
      return;
    }

    if (isBtcDenominated) {
      setBitcoinDenomination("sats");
      return;
    }

    setBtcValue(formatBtcInput(valueSat));
    setBitcoinDenomination("btc");
  }

  function handleChangeMode(event: React.ChangeEvent<HTMLInputElement>) {
    const nextValue = event.target.value.trim();

    if (mode === "bitcoin") {
      if (!isBtcDenominated && nextValue.includes(".")) {
        setBitcoinDenomination("btc");
        setBtcValue(nextValue);
        toast("Switched to BTC for decimal amount");

        if (!nextValue) {
          onValueSatChange("");
          return;
        }

        const amountBtc = Number(nextValue);
        if (!Number.isFinite(amountBtc)) {
          onValueSatChange("");
          return;
        }

        onValueSatChange(
          Math.max(0, Math.round(amountBtc * SATS_PER_BTC)).toString()
        );
        return;
      }

      if (isBtcDenominated) {
        setBtcValue(nextValue);

        if (!nextValue) {
          onValueSatChange("");
          return;
        }

        const amountBtc = Number(nextValue);
        if (!Number.isFinite(amountBtc)) {
          onValueSatChange("");
          return;
        }

        onValueSatChange(
          Math.max(0, Math.round(amountBtc * SATS_PER_BTC)).toString()
        );
        return;
      }

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

    if (!isFiatMode) {
      if (isBtcDenominated) {
        return amountSat / SATS_PER_BTC;
      }

      return amountSat;
    }

    if (!rate) {
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
      <InputGroup className="h-9 min-w-0 overflow-hidden has-[>[data-align=inline-start]]:[&>input]:pl-1">
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
          onChange={handleChangeMode}
          placeholder={
            isFiatMode ? "0.00" : isBtcDenominated ? "0.00000000" : "0"
          }
          required={required}
          step={isFiatMode ? "any" : isBtcDenominated ? 0.00000001 : 1}
          type="number"
          value={inputValue}
        />
        <InputGroupAddon align="inline-start">
          {isFiatMode ? (
            <InputGroupButton
              aria-label="Enter amount in bitcoin"
              disabled={disabled}
              onClick={handleToggleMode}
              size="xs"
              className="h-full rounded-none bg-transparent pl-2 pr-0 text-muted-foreground hover:bg-transparent hover:text-foreground"
              title="Enter amount in bitcoin"
            >
              {getCurrencySymbol(currency)}
            </InputGroupButton>
          ) : (
            <InputGroupButton
              aria-label={
                isBtcDenominated
                  ? "Display bitcoin amounts in satoshis"
                  : "Display bitcoin amounts in BTC"
              }
              aria-pressed={isBtcDenominated}
              disabled={disabled}
              onClick={handleToggleBitcoinDenomination}
              size="xs"
              className="h-full rounded-none bg-transparent pl-2 pr-0 text-muted-foreground hover:bg-transparent hover:text-foreground"
              title={
                isBtcDenominated
                  ? "Display bitcoin amounts in satoshis"
                  : "Display bitcoin amounts in BTC"
              }
            >
              {isBtcDenominated ? "BTC" : bitcoinUnit}
            </InputGroupButton>
          )}
        </InputGroupAddon>
        <InputGroupAddon
          align="inline-end"
          className="!mr-0 min-w-0 self-stretch py-0 pr-1"
        >
          {isFiatMode ? (
            <InputGroupButton
              aria-label={
                isBtcDenominated
                  ? "Display bitcoin amounts in satoshis"
                  : "Display bitcoin amounts in BTC"
              }
              aria-pressed={isBtcDenominated}
              disabled={disabled}
              onClick={handleToggleBitcoinDenomination}
              size="xs"
              className="sensitive slashed-zero h-full min-w-0 justify-end truncate rounded-none bg-transparent px-0.5 text-muted-foreground tabular-nums hover:bg-transparent hover:text-foreground"
              title={
                isBtcDenominated
                  ? "Display bitcoin amounts in satoshis"
                  : "Display bitcoin amounts in BTC"
              }
            >
              {alternateBitcoinValue.unit === "₿" && (
                <span>{alternateBitcoinValue.unit}</span>
              )}
              <span className="min-w-0 truncate">
                {alternateBitcoinValue.amount}
              </span>
              {alternateBitcoinValue.unit !== "₿" && (
                <span>{alternateBitcoinValue.unit}</span>
              )}
            </InputGroupButton>
          ) : (
            <InputGroupButton
              aria-label="Enter amount in fiat"
              disabled={disabled || !canUseFiat}
              onClick={handleAlternateValueClick}
              size="xs"
              className="sensitive slashed-zero h-full min-w-0 max-w-28 justify-end truncate rounded-none bg-transparent px-1 text-muted-foreground tabular-nums hover:bg-transparent hover:text-foreground sm:max-w-none"
              title="Enter amount in fiat"
            >
              {alternateValue ?? <Skeleton className="h-4 w-16" />}
            </InputGroupButton>
          )}
        </InputGroupAddon>
      </InputGroup>
      {!!contextRows?.length && (
        <div className="flex min-w-0 cursor-default flex-col gap-1 text-sm text-muted-foreground">
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
                    info?.bitcoinDisplayFormat,
                    bitcoinDenomination
                  )}
              </span>
            </div>
          ))}
        </div>
      )}
      {description && <FieldDescription>{description}</FieldDescription>}
      {error && <FieldError>{error}</FieldError>}
    </Field>
  );
}
