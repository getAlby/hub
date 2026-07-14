"use client";

import { WalletIcon } from "lucide-react";
import React from "react";
import AppAvatar from "src/components/AppAvatar";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  useComboboxAnchor,
} from "src/components/ui/combobox";
import { InputGroupAddon } from "src/components/ui/input-group";
import { Label } from "src/components/ui/label";
import { APP_SELECT_APPS_LIMIT } from "src/constants";
import { useApps } from "src/hooks/useApps";
import { getAppDisplayName } from "src/lib/utils";
import { App } from "src/types";

const LIGHTNING_BALANCE = "lightning-balance";
const LIGHTNING_BALANCE_LABEL = "Lightning Balance";

type ReceiveToOption = {
  value: string;
  label: string;
  app?: App;
};

type Props = {
  appId?: number;
  onChange(appId: number | undefined): void;
};

function LightningOption() {
  return (
    <div className="flex items-center gap-3">
      <div className="flex size-6 items-center justify-center rounded-lg bg-muted">
        <WalletIcon className="size-3 text-muted-foreground" />
      </div>
      <div>{LIGHTNING_BALANCE_LABEL}</div>
    </div>
  );
}

function AppOption({ app }: { app: App }) {
  return (
    <div className="flex items-center gap-3">
      <AppAvatar app={app} className="size-6 rounded-lg" />
      <div className="min-w-0">
        <div>{getAppDisplayName(app.name)}</div>
      </div>
    </div>
  );
}

export default function ReceiveToSelect({ appId, onChange }: Props) {
  const anchorRef = useComboboxAnchor();
  const [search, setSearch] = React.useState("");
  const { data: appsData } = useApps(APP_SELECT_APPS_LIMIT, undefined, {
    name: search,
  });

  const apps = React.useMemo(
    () =>
      [...(appsData?.apps || [])].sort((a, b) =>
        getAppDisplayName(a.name).localeCompare(getAppDisplayName(b.name))
      ),
    [appsData?.apps]
  );

  const options = React.useMemo<ReceiveToOption[]>(
    () => [
      { value: LIGHTNING_BALANCE, label: LIGHTNING_BALANCE_LABEL },
      ...apps.map((app) => ({
        value: app.id.toString(),
        label: getAppDisplayName(app.name),
        app,
      })),
    ],
    [apps]
  );

  const selectedOption = options.find((opt) =>
    appId ? opt.value === appId.toString() : undefined
  );

  return (
    <div className="grid w-full gap-1.5">
      <Label>Receive to</Label>
      <Combobox
        items={options}
        value={selectedOption}
        itemToStringValue={(option) => option.value}
        onInputValueChange={setSearch}
        onValueChange={(option) =>
          onChange(
            !option || option.value === LIGHTNING_BALANCE
              ? undefined
              : Number(option.value)
          )
        }
      >
        <div ref={anchorRef} className="w-full">
          <ComboboxInput placeholder={LIGHTNING_BALANCE_LABEL}>
            <InputGroupAddon>
              {selectedOption?.app ? (
                <AppAvatar
                  app={selectedOption.app}
                  className="size-6 rounded-full"
                />
              ) : (
                <div className="flex size-6 items-center justify-center rounded-sm bg-muted">
                  <WalletIcon className="size-3 text-muted-foreground" />
                </div>
              )}
            </InputGroupAddon>
          </ComboboxInput>
        </div>
        <ComboboxContent anchor={anchorRef}>
          <ComboboxEmpty>No connections found.</ComboboxEmpty>
          <ComboboxList>
            {(option: ReceiveToOption) => (
              <ComboboxItem key={option.value} value={option}>
                {option.app ? (
                  <AppOption app={option.app} />
                ) : (
                  <LightningOption />
                )}
              </ComboboxItem>
            )}
          </ComboboxList>
        </ComboboxContent>
      </Combobox>
    </div>
  );
}
