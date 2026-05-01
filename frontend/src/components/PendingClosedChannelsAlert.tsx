import { ExternalLinkIcon, HourglassIcon } from "lucide-react";
import ExternalLink from "src/components/ExternalLink";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { useInfo } from "src/hooks/useInfo";
import { useNodeDetails } from "src/hooks/useNodeDetails";
import { OnchainBalanceResponse, PendingBalancesDetails } from "src/types";

type PendingClosedChannelsAlertProps = {
  balance: OnchainBalanceResponse;
};

export function PendingClosedChannelsAlert({
  balance,
}: PendingClosedChannelsAlertProps) {
  if (balance.pendingBalancesFromChannelClosuresSat <= 0) {
    return null;
  }

  const pendingDetails = [
    ...balance.pendingBalancesDetails,
    ...balance.pendingSweepBalancesDetails,
  ];

  return (
    <Alert>
      <HourglassIcon />
      <AlertTitle>Pending Closed Channels</AlertTitle>
      <AlertDescription className="block">
        You have{" "}
        <FormattedBitcoinAmount
          amountMsat={balance.pendingBalancesFromChannelClosuresSat * 1000}
        />{" "}
        pending from closed channels
        {pendingDetails.length > 0 && (
          <>
            {" "}
            with
            {pendingDetails.map((details, index) => (
              <PendingBalancesDetailsItem
                key={`${details.fundingTxId}:${details.fundingTxVout}`}
                details={details}
                showSeparator={index < pendingDetails.length - 1}
              />
            ))}
          </>
        )}
        . Once spendable again these will become available in your on-chain
        balance. Funds from channels that were force closed may take up to 2
        weeks to become available.{" "}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/alby-hub/faq/why-was-my-lightning-channel-closed-and-what-to-do-next"
          className="underline"
        >
          Learn more
        </ExternalLink>
      </AlertDescription>
    </Alert>
  );
}

type PendingBalancesDetailsItemProps = {
  details: PendingBalancesDetails;
  showSeparator: boolean;
};

function PendingBalancesDetailsItem({
  details,
  showSeparator,
}: PendingBalancesDetailsItemProps) {
  const { data: info } = useInfo();
  const { data: nodeDetails } = useNodeDetails(details.nodeId);

  return (
    <span className="inline">
      &nbsp;
      <ExternalLink
        to={`https://amboss.space/node/${details.nodeId}`}
        className="underline"
      >
        {nodeDetails?.alias || "Unknown"}
        <ExternalLinkIcon className="ml-1 inline h-4 w-4" />
      </ExternalLink>{" "}
      (<FormattedBitcoinAmount amountMsat={details.amountSat * 1000} />
      )&nbsp;
      <ExternalLink
        to={`${info?.mempoolUrl}/tx/${details.fundingTxId}#flow=&vout=${details.fundingTxVout}`}
        className="underline"
      >
        funding tx
        <ExternalLinkIcon className="ml-1 inline h-4 w-4" />
      </ExternalLink>
      {showSeparator && ","}
    </span>
  );
}
