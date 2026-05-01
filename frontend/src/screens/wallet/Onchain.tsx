import { ArrowDownIcon, ArrowDownUpIcon, ArrowUpIcon } from "lucide-react";
import { useNavigate } from "react-router";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import FormattedFiatAmount from "src/components/FormattedFiatAmount";
import { OnchainTransactionsList } from "src/components/OnchainTransactionsList";
import { LinkButton } from "src/components/ui/custom/link-button";
import { useBalances } from "src/hooks/useBalances";

export default function Onchain() {
  const { data: balances } = useBalances(true);
  const navigate = useNavigate();

  if (!balances) {
    return null;
  }

  return (
    <>
      <div className="flex w-full flex-col items-center gap-8 pt-12 pb-16 text-center">
        <div className="flex flex-col items-center gap-4">
          <button
            type="button"
            onClick={() => navigate("/wallet")}
            aria-label="Toggle balance mode, currently On-chain Balance"
            className="inline-flex items-center justify-center gap-1 text-xs font-medium leading-none uppercase text-muted-foreground transition-colors hover:text-foreground"
          >
            On-chain Balance
            <ArrowDownUpIcon aria-hidden className="size-3 shrink-0" />
          </button>
          <div className="flex flex-col items-center gap-3">
            <div className="text-5xl md:text-6xl font-medium balance sensitive slashed-zero leading-none">
              <FormattedBitcoinAmount
                amountMsat={balances.onchain.spendableSat * 1000}
              />
            </div>
            <FormattedFiatAmount
              className="text-3xl font-normal leading-9 text-muted-foreground"
              amountSat={balances.onchain.spendableSat}
            />
          </div>
        </div>
        <div className="grid w-full max-w-100 grid-cols-2 items-center gap-3">
          <LinkButton to="/wallet/receive?type=onchain" size="lg">
            <ArrowDownIcon />
            Receive
          </LinkButton>
          <LinkButton to="/wallet/send" size="lg">
            <ArrowUpIcon />
            Send
          </LinkButton>
        </div>
      </div>
      <OnchainTransactionsList />
    </>
  );
}
