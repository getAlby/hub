import { AlertTriangleIcon } from "lucide-react";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "src/components/ui/tooltip";
import { cn } from "src/lib/utils";
import { OnchainBalanceResponse } from "src/types";

type OnchainBalanceSummaryProps = {
  balance: OnchainBalanceResponse;
  hasChannels: boolean;
  className?: string;
  amountClassName?: string;
  amountRowClassName?: string;
  fiatClassName?: string;
  incomingClassName?: string;
};

export function OnchainBalanceSummary({
  balance,
  hasChannels,
  className,
  amountClassName,
  amountRowClassName,
  fiatClassName,
  incomingClassName,
}: OnchainBalanceSummaryProps) {
  const showReserveWarning =
    hasChannels && balance.reserved + balance.spendable < 25_000;
  const incomingAmount = balance.total - balance.spendable;

  return (
    <div className={cn("flex flex-col gap-3", className)}>
      <div className={cn(amountRowClassName)}>
        <span
          className={cn(
            "mr-1 text-xl font-medium balance sensitive",
            amountClassName
          )}
        >
          <FormattedBitcoinAmount amount={balance.spendable * 1000} />
        </span>
        {showReserveWarning && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger>
                <AlertTriangleIcon className="size-3 text-muted-foreground" />
              </TooltipTrigger>
              <TooltipContent>
                You have insufficient funds in reserve to close channels or bump
                on-chain transactions and currently rely on the counterparty. It
                is recommended to deposit at least{" "}
                <FormattedBitcoinAmount amount={25_000 * 1000} /> to your
                on-chain balance.
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
      </div>
      <FormattedFiatAmount
        amount={balance.spendable}
        className={fiatClassName}
      />
      {incomingAmount > 0 && (
        <p
          className={cn(
            "text-xs text-muted-foreground animate-pulse",
            incomingClassName
          )}
        >
          +<FormattedBitcoinAmount amount={incomingAmount * 1000} /> incoming
        </p>
      )}
    </div>
  );
}
