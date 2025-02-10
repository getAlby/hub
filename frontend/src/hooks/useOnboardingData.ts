// src/hooks/useOnboardingData.ts

import { SUPPORT_ALBY_CONNECTION_NAME } from "src/constants";
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
  disabled: boolean;
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
  const { data: transactions } = useTransactions(undefined, false, 1);

  const isLoading =
    !apps ||
    !channels ||
    !info ||
    !nodeConnectionInfo ||
    !transactions ||
    (info.albyAccountConnected && (!albyMe || !albyBalance));

  if (isLoading) {
    return { isLoading: true, checklistItems: [] };
  }

  const isLinked =
    !!albyMe &&
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
  const hasSetupSupportPayment =
    apps &&
    apps.find((x) => x.name === SUPPORT_ALBY_CONNECTION_NAME) !== undefined;

  const checklistItems: Omit<ChecklistItem, "disabled">[] = [
    {
      title: "Open your first channel",
      description:
        "Establish a new Lightning channel to enable fast and low-fee Bitcoin transactions.",
      checked: hasChannel,
      to: "/channels/first",
    },
    ...(info.albyAccountConnected
      ? [
          {
            title: "Link to your Alby Account",
            description:
              "Link your lightning address & other apps to this Hub.",
            checked: isLinked,
            to: "/apps",
          },
        ]
      : []),
    {
      title: "Send or receive your first payment",
      description:
        "Use your newly opened channel to make a transaction on the Lightning Network.",
      checked: hasTransaction,
      to: "/wallet",
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
            to: "/settings/backup",
          },
        ]
      : []),
    ...(!info.oauthRedirect
      ? [
          {
            title: "Support Alby Hub",
            description:
              "Setup a recurring payment to support the development of Alby Hub",
            checked: hasSetupSupportPayment,
            to: "/support-alby",
          },
        ]
      : []),
  ];

  const nextStep = checklistItems.find((x) => !x.checked);

  const sortedChecklistItems = checklistItems.map((item) => ({
    ...item,
    disabled: item !== nextStep,
  }));

  return { isLoading: false, checklistItems: sortedChecklistItems };
};
