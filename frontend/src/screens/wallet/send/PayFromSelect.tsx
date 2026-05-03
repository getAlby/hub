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
import { PAY_FROM_SELECT_APPS_LIMIT } from "src/constants";
import { useApps } from "src/hooks/useApps";
import { getAppDisplayName } from "src/lib/utils";
import { App } from "src/types";

const SPENDING_BALANCE = "spending-balance";
const SPENDING_BALANCE_LABEL = "Spending Balance";

type PayFromOption = {
  value: string;
  label: string;
  app?: App;
};

type Props = {
  appId?: number;
  onChange(appId: number | undefined): void;
};

function SpendingOption() {
  return (
    <div className="flex items-center gap-3">
      <div className="flex size-6 items-center justify-center rounded-lg bg-muted">
        <WalletIcon className="size-3 text-muted-foreground" />
      </div>
      <div>{SPENDING_BALANCE_LABEL}</div>
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

export default function PayFromSelect({ appId, onChange }: Props) {
  const anchorRef = useComboboxAnchor();
  const [search, setSearch] = React.useState("");
  const { data: appsData } = useApps(PAY_FROM_SELECT_APPS_LIMIT, undefined, {
    name: search,
  });

  const apps = React.useMemo(
    () =>
      [...(appsData?.apps || [])]
        .filter((app) => app.scopes.includes("pay_invoice"))
        .sort((a, b) =>
          getAppDisplayName(a.name).localeCompare(getAppDisplayName(b.name))
        ),
    [appsData?.apps]
  );

  const options = React.useMemo<PayFromOption[]>(
    () => [
      { value: SPENDING_BALANCE, label: SPENDING_BALANCE_LABEL },
      ...apps.map((app) => ({
        value: app.id.toString(),
        label: getAppDisplayName(app.name),
        app,
      })),
    ],
    [apps]
  );

  const selectedOption = options.find(
    (opt) => (appId ? opt.value === appId.toString() : undefined) // : opt.value === SPENDING_BALANCE
  );

  return (
    <div className="grid w-full gap-1.5">
      <Label>Pay from</Label>
      <Combobox
        items={options}
        value={selectedOption}
        itemToStringValue={(option) => option.value}
        onInputValueChange={setSearch}
        onValueChange={(option) =>
          onChange(
            option?.value === SPENDING_BALANCE
              ? undefined
              : Number(option?.value)
          )
        }
      >
        <div ref={anchorRef} className="w-full">
          <ComboboxInput placeholder={SPENDING_BALANCE_LABEL}>
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
            {(option: PayFromOption) => (
              <ComboboxItem key={option.value} value={option}>
                {option.app ? (
                  <AppOption app={option.app} />
                ) : (
                  <SpendingOption />
                )}
              </ComboboxItem>
            )}
          </ComboboxList>
        </ComboboxContent>
      </Combobox>
    </div>
  );
}
