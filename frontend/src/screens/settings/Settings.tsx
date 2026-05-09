import {
  CheckIcon,
  LockIcon,
  MonitorIcon,
  MoonIcon,
  StarsIcon,
  SunIcon,
} from "lucide-react";
import React from "react";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { ThemePreview } from "src/components/ThemePreview";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { Badge } from "src/components/ui/badge";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { Separator } from "src/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "src/components/ui/tabs";
import { DarkMode, Themes, useTheme } from "src/components/ui/theme-provider";
import {
  BITCOIN_DISPLAY_FORMAT_BIP177,
  BITCOIN_DISPLAY_FORMAT_SATS,
} from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCurrencies } from "src/hooks/useCurrencies";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function Settings() {
  const { data: albyMe } = useAlbyMe();
  const { theme, darkMode, setTheme, setDarkMode } = useTheme();
  const { currencies, isLoading: isCurrenciesLoading } = useCurrencies();
  const [showUpgradeDialog, setShowUpgradeDialog] = React.useState(false);

  const { data: info, mutate: reloadInfo } = useInfo();

  async function updateSettings(
    payload: Record<string, string | boolean>,
    successMessage: string,
    errorMessage: string
  ) {
    try {
      await request("/api/settings", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      await reloadInfo();
      toast(successMessage);
    } catch (error) {
      console.error(error);
      handleRequestError(errorMessage, error);
    }
  }

  async function updateCurrency(currency: string) {
    await updateSettings(
      { currency },
      `Currency set to ${currency}`,
      "Failed to update currencies"
    );
  }

  async function updateBitcoinDisplayFormat(bitcoinDisplayFormat: string) {
    await updateSettings(
      { bitcoinDisplayFormat },
      "Bitcoin display format updated",
      "Failed to update bitcoin display format"
    );
  }

  if (!info) {
    return <Loading />;
  }

  const paidThemes = ["matrix", "ghibli", "claymorphism"];
  const hasPlan = !!albyMe?.subscription.plan_code;

  const darkModeOptions: {
    value: DarkMode;
    icon: React.ReactNode;
    label: string;
  }[] = [
    { value: "light", icon: <SunIcon className="size-4" />, label: "Light" },
    { value: "dark", icon: <MoonIcon className="size-4" />, label: "Dark" },
    {
      value: "system",
      icon: <MonitorIcon className="size-4" />,
      label: "System",
    },
  ];

  return (
    <>
      <SettingsHeader
        pageTitle="Settings"
        title="General"
        description="Customize how Alby Hub looks and feels."
      />
      <div className="flex flex-col gap-6 pb-10">
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1 text-sm">
            <h3 className="font-semibold">Appearance</h3>
            <p className="text-muted-foreground">
              Choose a theme and light/dark mode.
            </p>
          </div>
          <div className="space-y-6">
            <div className="space-y-3">
              <Label id="theme-label">Theme</Label>
              <div
                role="radiogroup"
                aria-labelledby="theme-label"
                className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3"
              >
                {Themes.map((t) => {
                  const isPaidTheme = paidThemes.includes(t);
                  const isDisabled = isPaidTheme && !hasPlan;
                  const isSelected = theme === t;

                  return (
                    <button
                      key={t}
                      type="button"
                      role="radio"
                      aria-checked={isSelected}
                      onClick={() => {
                        if (isDisabled) {
                          setShowUpgradeDialog(true);
                          return;
                        }
                        setTheme(t);
                        toast("Theme updated.");
                      }}
                      className={cn(
                        "group relative flex flex-col rounded-lg border-2 text-left transition-all w-full overflow-hidden hover:border-primary/50",
                        isSelected
                          ? "border-primary ring-2 ring-primary/20"
                          : "border-border",
                        isDisabled
                          ? "cursor-not-allowed opacity-60 hover:border-border"
                          : "cursor-pointer"
                      )}
                    >
                      <ThemePreview theme={t} />
                      <div className="flex items-center justify-center gap-1.5 py-1.5 px-1 w-full">
                        <span className="text-xs font-medium capitalize truncate">
                          {t}
                        </span>
                        {isPaidTheme && (
                          <Badge
                            variant="outline"
                            className="text-[10px] px-1 py-0"
                          >
                            <StarsIcon className="size-2.5" />
                            Pro
                          </Badge>
                        )}
                      </div>
                      {isSelected && (
                        <div className="absolute top-1.5 right-1.5 size-4 rounded-full bg-primary flex items-center justify-center">
                          <CheckIcon className="size-2.5 text-primary-foreground" />
                        </div>
                      )}
                      {isDisabled && (
                        <div className="absolute top-1.5 right-1.5 size-4 rounded-full bg-background flex items-center justify-center">
                          <LockIcon className="size-2.5 text-foreground" />
                        </div>
                      )}
                    </button>
                  );
                })}
              </div>
              <UpgradeDialog
                open={showUpgradeDialog}
                onOpenChange={setShowUpgradeDialog}
              />
            </div>

            <div className="space-y-3">
              <Label id="dark-mode-label">Mode</Label>
              <Tabs value={darkMode}>
                <TabsList>
                  {darkModeOptions.map((option) => (
                    <TabsTrigger
                      value={option.value}
                      onClick={() => {
                        setDarkMode(option.value);
                        toast("Appearance updated.");
                      }}
                      className="px-3"
                    >
                      {option.icon}
                      {option.label}
                    </TabsTrigger>
                  ))}
                </TabsList>
              </Tabs>
            </div>
          </div>
        </div>
        <Separator />
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1 text-sm">
            <h3 className="font-semibold">Units & Currency</h3>
            <p className="text-muted-foreground">
              Choose how amounts are displayed.
            </p>
          </div>
          <div className="space-y-4">
            <div className="grid gap-1.5">
              <Label htmlFor="bitcoinDisplayFormat">Display Unit</Label>
              <Select
                value={info.bitcoinDisplayFormat}
                onValueChange={updateBitcoinDisplayFormat}
              >
                <SelectTrigger className="w-full md:w-60">
                  <SelectValue placeholder="Select a display format" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value={BITCOIN_DISPLAY_FORMAT_BIP177}>
                    ₿
                  </SelectItem>
                  <SelectItem value={BITCOIN_DISPLAY_FORMAT_SATS}>
                    sats
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-1.5">
              <Label htmlFor="currency">Fiat Currency</Label>
              <Select
                value={info?.currency}
                onValueChange={updateCurrency}
                disabled={isCurrenciesLoading}
              >
                <SelectTrigger className="w-full md:w-60">
                  <SelectValue
                    placeholder={
                      isCurrenciesLoading
                        ? "Loading currencies..."
                        : "Select a currency"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  {currencies.map(([code, name]) => (
                    <SelectItem key={code} value={code}>
                      {name} ({code})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}

export default Settings;
