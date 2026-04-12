import {
  CheckIcon,
  CoinsIcon,
  LockIcon,
  MonitorIcon,
  MoonIcon,
  PaletteIcon,
  StarsIcon,
  SunIcon,
} from "lucide-react";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import { UpgradeDialog } from "src/components/UpgradeDialog";
import { Badge } from "src/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
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
  const { theme, darkMode, isDarkMode, setTheme, setDarkMode } = useTheme();
  const { currencies, isLoading: isCurrenciesLoading } = useCurrencies();

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
      <title>Settings · Alby Hub</title>
      <div className="w-full flex flex-col gap-6">
        {/* Theme & Appearance Section */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <PaletteIcon className="size-5 text-muted-foreground" />
              <CardTitle>Appearance</CardTitle>
            </div>
            <CardDescription>
              Customize how Alby Hub looks and feels.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="space-y-3">
              <Label>Theme</Label>
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
                {Themes.map((t) => {
                  const isPaidTheme = paidThemes.includes(t);
                  const isDisabled = isPaidTheme && !hasPlan;
                  const isSelected = theme === t;
                  const previewClass = cn(`theme-${t}`, isDarkMode && "dark");

                  const themeCard = (
                    <button
                      key={t}
                      type="button"
                      aria-pressed={isSelected}
                      aria-disabled={isDisabled || undefined}
                      onClick={
                        isDisabled
                          ? undefined
                          : () => {
                              setTheme(t);
                              toast("Theme updated.");
                            }
                      }
                      className={cn(
                        "group relative flex flex-col rounded-lg border-2 p-1 text-left transition-all cursor-pointer w-full",
                        "hover:border-primary/50",
                        isSelected
                          ? "border-primary ring-2 ring-primary/20"
                          : "border-border",
                        isDisabled && "opacity-50 hover:border-border"
                      )}
                    >
                      {/* Mini preview using actual theme CSS variables */}
                      <div
                        className={cn(
                          "rounded-md w-full aspect-16/10 flex flex-col overflow-hidden bg-background",
                          previewClass
                        )}
                      >
                        <div className="h-2 w-full bg-primary" />
                        <div className="flex-1 p-1.5 flex flex-col gap-1">
                          <div className="h-1 w-3/4 rounded-full bg-muted" />
                          <div className="h-1 w-1/2 rounded-full bg-muted" />
                          <div className="mt-auto flex gap-1">
                            <div className="h-1.5 w-6 rounded-full bg-primary" />
                            <div className="h-1.5 w-4 rounded-full bg-muted" />
                          </div>
                        </div>
                      </div>
                      {/* Label */}
                      <div className="flex items-center justify-center gap-1.5 py-1.5 px-1">
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
                      {/* Selected indicator */}
                      {isSelected && (
                        <div className="absolute top-1.5 right-1.5 size-4 rounded-full bg-primary flex items-center justify-center">
                          <CheckIcon className="size-2.5 text-primary-foreground" />
                        </div>
                      )}
                      {/* Locked indicator */}
                      {isDisabled && (
                        <div className="absolute inset-0 rounded-lg flex items-center justify-center bg-background/50">
                          <LockIcon className="size-4 text-muted-foreground" />
                        </div>
                      )}
                    </button>
                  );

                  if (isDisabled) {
                    return <UpgradeDialog key={t}>{themeCard}</UpgradeDialog>;
                  }

                  return themeCard;
                })}
              </div>
            </div>

            <div className="space-y-3">
              <Label>Appearance</Label>
              <div className="inline-flex rounded-lg border bg-muted p-1 gap-1">
                {darkModeOptions.map((option) => (
                  <button
                    key={option.value}
                    type="button"
                    aria-pressed={darkMode === option.value}
                    onClick={() => {
                      setDarkMode(option.value);
                      toast("Appearance updated.");
                    }}
                    className={cn(
                      "inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-all",
                      darkMode === option.value
                        ? "bg-background text-foreground shadow-sm"
                        : "text-muted-foreground hover:text-foreground"
                    )}
                  >
                    {option.icon}
                    {option.label}
                  </button>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Units & Currency Section */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <CoinsIcon className="size-5 text-muted-foreground" />
              <CardTitle>Units & Currency</CardTitle>
            </div>
            <CardDescription>Choose how amounts are displayed.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
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
          </CardContent>
        </Card>
      </div>
    </>
  );
}

export default Settings;
