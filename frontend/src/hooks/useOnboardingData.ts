// src/hooks/useOnboardingData.ts

import { ALBY_MIN_BALANCE, ALBY_SERVICE_FEE } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { useTransactions } from "src/hooks/useTransactions";

interface ChecklistItem {
  title: string;
  description: string;
  checked: boolean;
  to: string;
}

interface UseOnboardingDataResponse {
  isLoading: boolean;
  checklistItems: ChecklistItem[];
}

export const useOnboardingData = (): UseOnboardingDataResponse => {
  const { data: albyBalance } = useAlbyBalance();
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
    !transactions ||
    !albyBalance;

  if (isLoading) {
    return { isLoading: true, checklistItems: [] };
  }

  const isLinked =
    albyMe &&
    nodeConnectionInfo &&
    albyMe?.keysend_pubkey === nodeConnectionInfo?.pubkey;
  const hasChannel =
    !hasChannelManagement || (hasChannelManagement && channels.length > 0);
  const hasBackedUp =
    hasMnemonic === true &&
    info &&
    info.nextBackupReminder !== "" &&
    new Date(info.nextBackupReminder).getTime() > new Date().getTime();
  const hasCustomApp =
    apps && apps.find((x) => x.name !== "getalby.com") !== undefined;
  const hasTransaction = transactions.length > 0;

  const canMigrateAlbyFundsToNewChannel =
    hasChannelManagement &&
    info.backendType === "LDK" &&
    albyBalance.sats * (1 - ALBY_SERVICE_FEE) >
      ALBY_MIN_BALANCE + 50000; /* accommodate for onchain fees */

  const checklistItems: ChecklistItem[] = [
    {
      title: "Open your first channel",
      description:
        "Establish a new Lightning channel to enable fast and low-fee Bitcoin transactions.",
      checked: hasChannel,
      to: canMigrateAlbyFundsToNewChannel
        ? "/onboarding/lightning/migrate-alby"
        : "/channels/outgoing",
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
            checked: hasBackedUp === true,
            to: "/settings/key-backup",
          },
        ]
      : []),
  ];

  const sortedChecklistItems = checklistItems.sort(
    (a, b) => Number(b.checked) - Number(a.checked)
  );

  return { isLoading: false, checklistItems: sortedChecklistItems };
};
