import {
  CreditCard,
  FileText,
  HelpCircle,
  Home,
  Info,
  Network,
  Plug,
  QrCode,
  Send,
  Settings,
  Shield,
  Shuffle,
  Store,
  User,
  Users,
  Wallet,
  Zap,
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
          <CommandItem onSelect={() => runCommand(() => navigate("/home"))}>
            <Home />
            <span>Dashboard</span>
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
            <span>Connections</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/sub-wallets"))}
          >
            <CreditCard />
            <span>Sub-wallets</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => navigate("/channels"))}>
            <Network />
            <span>Channels</span>
          </CommandItem>
          <CommandItem onSelect={() => runCommand(() => navigate("/peers"))}>
            <Users />
            <span>Peers</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/apps?tab=app-store"))}
          >
            <Store />
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
            <Zap />
            <span>Receive On-chain</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/wallet/sign-message"))}
          >
            <Shield />
            <span>Sign Message</span>
          </CommandItem>
        </CommandGroup>
        <CommandSeparator />
        <CommandGroup heading="Settings">
          <CommandItem onSelect={() => runCommand(() => navigate("/settings"))}>
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
            onSelect={() => runCommand(() => navigate("/settings/about"))}
          >
            <Info />
            <span>About</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/settings/developer"))}
          >
            <HelpCircle />
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
            onSelect={() => runCommand(() => navigate("/sub-wallets/new"))}
          >
            <CreditCard />
            <span>Create Sub-wallet</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/peers/new"))}
          >
            <Users />
            <span>Connect Peer</span>
          </CommandItem>
          <CommandItem
            onSelect={() => runCommand(() => navigate("/channels/first"))}
          >
            <Network />
            <span>Open First Channel</span>
          </CommandItem>
        </CommandGroup>
        {!!connectedAppsByAppName?.apps.length && (
          <CommandGroup heading="Connected Apps" forceMount>
            {connectedAppsByAppName?.apps.map((app) => (
              <CommandItem
                key={app.id}
                onSelect={() => runCommand(() => navigate(`/apps/${app.id}`))}
              >
                <AppAvatar app={app} className="w-4 h-4" />
                <span>{app.name}</span>
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}
