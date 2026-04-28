import { WalletIcon } from "lucide-react";
import React from "react";
import AppAvatar from "src/components/AppAvatar";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useApps } from "src/hooks/useApps";
import { getAppDisplayName } from "src/lib/utils";
import { App } from "src/types";

const MAIN_WALLET = "main-wallet";

type Props = {
  appId?: number;
  onChange(appId: number | undefined): void;
};

function getOptionValue(appId?: number) {
  return appId ? appId.toString() : MAIN_WALLET;
}

function MainWalletOption() {
  return (
    <div className="flex items-center gap-3">
      <div className="flex size-7 items-center justify-center rounded-lg bg-muted">
        <WalletIcon className="size-3 text-muted-foreground" />
      </div>
      <div>Main wallet</div>
    </div>
  );
}

function AppOption({ app }: { app: App }) {
  return (
    <div className="flex items-center gap-3">
      <AppAvatar app={app} className="size-7 rounded-lg" />
      <div className="min-w-0">
        <div>{getAppDisplayName(app.name)}</div>
      </div>
    </div>
  );
}

export default function PayFromSelect({ appId, onChange }: Props) {
  const { data: appsData } = useApps(100, undefined, {
    subWallets: false,
  });

  const apps = React.useMemo(
    () =>
      (appsData?.apps || []).filter((app) =>
        app.scopes.includes("pay_invoice")
      ),
    [appsData?.apps]
  );

  const selectedApp = apps.find((app) => app.id === appId);

  return (
    <div className="grid gap-1.5 w-full">
      <Label>Pay from</Label>
      <Select
        value={getOptionValue(selectedApp?.id)}
        onValueChange={(value) =>
          onChange(value === MAIN_WALLET ? undefined : Number(value))
        }
      >
        <SelectTrigger className="w-full">
          <SelectValue placeholder="Select payment source">
            {selectedApp ? (
              <AppOption app={selectedApp} />
            ) : (
              <MainWalletOption />
            )}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value={MAIN_WALLET}>
            <MainWalletOption />
          </SelectItem>
          {apps.map((app) => (
            <SelectItem key={app.id} value={app.id.toString()}>
              <AppOption app={app} />
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
