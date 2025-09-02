import {
  ArrowUpDown,
  Code2,
  CreditCard,
  FileSignature,
  FileText,
  Home,
  Info,
  LayoutGrid,
  Link,
  Network,
  Plug,
  QrCode,
  RefreshCw,
  Send,
  Settings,
  Shield,
  Shuffle,
  SquareStack,
  User,
  UserPlus2,
  Wallet,
} from "lucide-react";
import * as React from "react";
import { useNavigate } from "react-router-dom";
import AppAvatar from "src/components/AppAvatar";

import {
  CommandDialog,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "src/components/ui/command";
import { useApps } from "src/hooks/useApps";

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

  const runCommand = React.useCallback(
    (command: () => void) => {
      onOpenChange?.(false);
      command();
      setSearchText("");
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
            <Home />
            <span>Home</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => navigate("/wallet"))}>
            <Wallet />
            <span>Wallet</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/apps?tab=connected-apps"))
            }
          >
            <Plug />
            <span>Connected Apps</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/sub-wallets"))}
          >
            <CreditCard />
            <span>Sub-wallets</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/channels"))}
            keywords={["node", "liquidity", "channels"]}
          >
            <Network />
            <span>Node</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => navigate("/peers"))}>
            <Network />
            <span>Peers</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/apps?tab=app-store"))}
          >
            <LayoutGrid />
            <span>App Store</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Wallet Actions">
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/send"))}
          >
            <Send />
            <span>Send Payment</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/receive"))}
          >
            <QrCode />
            <span>Receive Payment</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/swap"))}
          >
            <Shuffle />
            <span>Swap</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/swap/auto"))}
          >
            <RefreshCw />
            <span>Auto Swap</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/wallet/receive/invoice"))
            }
          >
            <FileText />
            <span>Create Invoice</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/wallet/receive/onchain"))
            }
          >
            <Link />
            <span>Receive On-chain</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/sign-message"))}
          >
            <FileSignature />
            <span>Sign Message</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Settings">
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings"))}
            keywords={["theme", "fiat", "currency", "dark"]}
          >
            <Settings />
            <span>Settings</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings/backup"))}
          >
            <Shield />
            <span>Backup</span>
          </CommandItem>
          <CommandItem
            onSelect={() =>
              runCommand(() => navigate("/settings/alby-account"))
            }
          >
            <User />
            <span>Alby Account</span>
          </CommandItem>
          <CommandItem
            keywords={["info"]}
            onSelect={() => runCommand(() => navigate("/settings/about"))}
          >
            <Info />
            <span>About</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings/developer"))}
          >
            <Code2 />
            <span>Developer Settings</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Quick Actions">
          <CommandItem onSelect={() => runCommand(() => navigate("/apps/new"))}>
            <Plug />
            <span>Connect New App</span>
          </CommandItem>
          <CommandItem
            keywords={["New Sub-Wallet"]}
            onSelect={() => runCommand(() => navigate("/sub-wallets/new"))}
          >
            <SquareStack />
            <span>Create Sub-wallet</span>
          </CommandItem>
          <CommandItem
            keywords={["New Channel"]}
            onSelect={() => runCommand(() => navigate("/channels/incoming"))}
          >
            <ArrowUpDown />
            <span>Open Channel</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/peers/new"))}
          >
            <UserPlus2 />
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
      </CommandList>
    </CommandDialog>
  );
}
