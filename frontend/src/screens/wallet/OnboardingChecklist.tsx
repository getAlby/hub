import { ChevronRight, Circle, CircleCheck } from "lucide-react";
import { Link } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { useTransactions } from "src/hooks/useTransactions";
import { cn } from "src/lib/utils";

function OnboardingChecklist() {
  // const { data: albyBalance } = useAlbyBalance();
  const { data: albyMe } = useAlbyMe();
  const { data: apps } = useApps();
  const { data: channels } = useChannels();
  const { data: info, hasChannelManagement, hasMnemonic } = useInfo();
  const { data: nodeConnectionInfo } = useNodeConnectionInfo();
  const { data: transactions } = useTransactions(false, 1);

  const isLoading =
    !albyMe ||
    !apps ||
    !channels ||
    !info ||
    !nodeConnectionInfo ||
    !transactions;

  if (isLoading) {
    return;
  }

  /*const hasAlbyBalance =
    hasChannelManagement &&
    albyBalance &&
    albyBalance.sats * (1 - ALBY_SERVICE_FEE) >
      ALBY_MIN_BALANCE + 50000; // accommodate for on-chain fees
      */

  const isLinked =
    albyMe &&
    nodeConnectionInfo &&
    albyMe?.keysend_pubkey === nodeConnectionInfo?.pubkey;
  const hasChannel =
    !hasChannelManagement || (hasChannelManagement && channels.length > 0);
  const hasBackedUp =
    hasMnemonic &&
    info &&
    info.nextBackupReminder &&
    new Date(info.nextBackupReminder).getTime() > new Date().getTime();
  const hasCustomApp =
    apps && apps.find((x) => x.name !== "getalby.com") !== undefined;
  const hasTransaction = transactions.length > 0;

  if (
    isLinked &&
    hasChannel &&
    (!hasMnemonic || hasBackedUp) &&
    hasCustomApp &&
    hasTransaction
  ) {
    return;
  }

  const checklistItems = [
    {
      title: "Open your first channel",
      description:
        "Establish a new Lightning channel to enable fast and low-fee Bitcoin transactions.",
      checked: hasChannel,
      to: "/channels",
    },
    {
      title: "Send or receive your first payment",
      description:
        "Use your newly opened channel to make a transaction on the Lightning Network.",
      checked: hasTransaction,
      to: "/wallet",
    },
    {
      title: "Link to your Alby Account",
      description: "Link your lightning address & other apps to this Hub.",
      checked: isLinked,
      to: "/apps",
    },
    // TODO: enable when we can always migrate funds
    /*{
      title: "Migrate your balance to your Hub",
      description: "Move your existing funds into self-custody.",
      checked: !hasAlbyBalance,
      to: "/onboarding/lightning/migrate-alby",
    },*/
    {
      title: "Connect your first app",
      description:
        "Seamlessly connect apps and integrate your wallet with other apps from your Hub.",
      checked: hasCustomApp,
      to: "/appstore",
    },
    ...(hasMnemonic
      ? [
          {
            title: "Backup your keys",
            description:
              "Secure your keys by creating a backup to ensure you don't lose access.",
            checked: hasBackedUp,
            to: "/settings/key-backup",
          },
        ]
      : []),
  ];

  const sortedChecklistItems = checklistItems.sort(
    (a, b) => (b && b.checked ? 1 : 0) - (a && a.checked ? 1 : 0)
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle>Get started with your Alby Hub</CardTitle>
        <CardDescription>
          Follow these initial steps to set up and make the most of your Alby
          Hub.
        </CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col">
        {sortedChecklistItems.map((item) => (
          <ChecklistItem
            key={item.title}
            title={item.title}
            description={item.description}
            checked={!!item.checked}
            to={item.to}
          />
        ))}
      </CardContent>
    </Card>
  );
}

type ChecklistItemProps = {
  title: string;
  checked: boolean;
  description: string;
  to: string;
};

function ChecklistItem({
  title,
  checked = false,
  description,
  to,
}: ChecklistItemProps) {
  const content = (
    <div
      className={cn(
        "flex flex-col p-3 relative group rounded-lg",
        !checked && "hover:bg-muted"
      )}
    >
      {!checked && (
        <div className="absolute top-0 left-0 w-full h-full items-center justify-end pr-1.5 hidden group-hover:flex opacity-25">
          <ChevronRight className="w-8 h-8" />
        </div>
      )}
      <div className="flex items-center gap-2">
        {checked ? (
          <CircleCheck className="w-5 h-5" />
        ) : (
          <Circle className="w-5 h-5" />
        )}
        <div
          className={cn(
            "text-sm font-medium leading-none",
            checked && "line-through"
          )}
        >
          {title}
        </div>
      </div>
      {!checked && (
        <div className="text-muted-foreground text-sm ml-7">{description}</div>
      )}
    </div>
  );

  return checked ? content : <Link to={to}>{content}</Link>;
}

export default OnboardingChecklist;
