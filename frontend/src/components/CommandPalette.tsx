import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import {
  ArrowDownIcon,
  ArrowUpDownIcon,
  ArrowUpIcon,
  Code2Icon,
  CreditCardIcon,
  FileSignatureIcon,
  FileTextIcon,
  HomeIcon,
  InfoIcon,
  LayoutGridIcon,
  LinkIcon,
  NetworkIcon,
  PlugIcon,
  QrCodeIcon,
  RefreshCwIcon,
  SendIcon,
  SettingsIcon,
  ShieldIcon,
  ShuffleIcon,
  SquareStackIcon,
  UserIcon,
  UserPlus2Icon,
  WalletIcon,
} from "lucide-react";
import * as React from "react";
import { useNavigate } from "react-router";
import useSWR from "swr";
import AppAvatar from "src/components/AppAvatar";
import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";

import {
  CommandDialog,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "src/components/ui/command";
import { useApps } from "src/hooks/useApps";
import { getTransactionsUrl } from "src/hooks/useTransactions";
import { ListTransactionsResponse } from "src/types";
import { swrFetcher } from "src/utils/swr";

dayjs.extend(relativeTime);

interface CommandPaletteProps {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  const navigate = useNavigate();
  const [searchText, setSearchText] = React.useState("");

  const { data: connectedAppsByAppName } = useApps(
    undefined,
    undefined,
    {
      name: searchText,
    },
    undefined,
    !!searchText?.length
  );

  const { data: transactionsBySearch } = useSWR<ListTransactionsResponse>(
    searchText ? getTransactionsUrl(undefined, 5, 1, searchText) : null,
    swrFetcher,
    { keepPreviousData: true }
  );

  const runCommand = React.useCallback(
    (command: () => void) => {
      onOpenChange?.(false);
      setSearchText("");
      // hack: run the command after 1ms delay to fix autofocus
      setTimeout(command, 1);
    },
    [onOpenChange]
  );

  return (
    <CommandDialog open={open} onOpenChange={onOpenChange}>
      <CommandInput
        placeholder="Type a command or search..."
        value={searchText}
        onValueChange={setSearchText}
      />

      <CommandList>
        {/* <CommandEmpty>No results found</CommandEmpty> */}
        <CommandGroup heading="Navigation">
          <CommandItem
            onSelect={() => runCommand(() => navigate("/home"))}
            keywords={["dashboard"]}
          >
            <HomeIcon />
            <span>Home</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => navigate("/wallet"))}>
            <WalletIcon />
            <span>Wallet</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/apps?tab=connected-apps"))
            }
          >
            <PlugIcon />
            <span>Connected Apps</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/sub-wallets"))}
          >
            <CreditCardIcon />
            <span>Sub-wallets</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/channels"))}
            keywords={["node", "liquidity", "channels"]}
          >
            <NetworkIcon />
            <span>Node</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => navigate("/peers"))}>
            <NetworkIcon />
            <span>Peers</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/apps?tab=app-store"))}
          >
            <LayoutGridIcon />
            <span>App Store</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Wallet Actions">
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/send"))}
          >
            <SendIcon />
            <span>Send Payment</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/receive"))}
          >
            <QrCodeIcon />
            <span>Receive Payment</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/swap"))}
          >
            <ShuffleIcon />
            <span>Swap</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/swap/auto"))}
          >
            <RefreshCwIcon />
            <span>Auto Swap</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/wallet/receive/invoice"))
            }
          >
            <FileTextIcon />
            <span>Create Invoice</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/wallet/receive?type=onchain"))
            }
          >
            <LinkIcon />
            <span>Receive On-chain</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/sign-message"))}
          >
            <FileSignatureIcon />
            <span>Sign Message</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Settings">
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings"))}
            keywords={["theme", "fiat", "currency", "dark"]}
          >
            <SettingsIcon />
            <span>Settings</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings/backup"))}
          >
            <ShieldIcon />
            <span>Backup</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/settings/alby-account"))
            }
          >
            <UserIcon />
            <span>Alby Account</span>
          </CommandItem>
          <CommandItem
            keywords={["info"]}
            onSelect={() => runCommand(() => navigate("/settings/about"))}
          >
            <InfoIcon />
            <span>About</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings/developer"))}
          >
            <Code2Icon />
            <span>Developer Settings</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Quick Actions">
          <CommandItem onSelect={() => runCommand(() => navigate("/apps/new"))}>
            <PlugIcon />
            <span>Connect New App</span>
          </CommandItem>
          <CommandItem
            keywords={["New Sub-Wallet"]}
            onSelect={() => runCommand(() => navigate("/sub-wallets/new"))}
          >
            <SquareStackIcon />
            <span>Create Sub-wallet</span>
          </CommandItem>
          <CommandItem
            keywords={["New Channel"]}
            onSelect={() => runCommand(() => navigate("/channels/incoming"))}
          >
            <ArrowUpDownIcon />
            <span>Open Channel</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/peers/new"))}
          >
            <UserPlus2Icon />
            <span>Connect Peer</span>
          </CommandItem>
        </CommandGroup>
        {!!connectedAppsByAppName?.apps.length && (
          <CommandGroup heading="Connected Apps" forceMount>
            {connectedAppsByAppName?.apps.map((app) => (
              <CommandItem
                key={app.id}
                onSelect={() => runCommand(() => navigate(`/apps/${app.id}`))}
              >
                <AppAvatar app={app} className="size-4" />
                <span>{app.name}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
        {!!searchText && !!transactionsBySearch?.transactions.length && (
          <CommandGroup heading="Transactions" forceMount>
            {transactionsBySearch.transactions.map((tx) => (
              <CommandItem
                key={tx.paymentHash}
                value={`tx-${tx.paymentHash}`}
                onSelect={() =>
                  runCommand(() => navigate(`/wallet?q=${tx.paymentHash}`))
                }
              >
                {tx.type === "incoming" ? <ArrowDownIcon /> : <ArrowUpIcon />}
                <span className="truncate sensitive">
                  {tx.description ||
                    (tx.type === "incoming"
                      ? tx.metadata?.payer_data?.name
                        ? `Received from ${tx.metadata.payer_data.name}`
                        : "Received"
                      : tx.metadata?.recipient_data?.identifier
                        ? `Sent to ${tx.metadata.recipient_data.identifier}`
                        : "Sent")}
                </span>
                <span className="ml-auto flex shrink-0 items-center gap-2 slashed-zero">
                  <span className="sensitive">
                    {tx.type === "outgoing" ? "-" : "+"}
                    <FormattedBitcoinAmount amountMsat={tx.amountMsat} />
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {dayjs(tx.updatedAt).fromNow()}
                  </span>
                </span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}
