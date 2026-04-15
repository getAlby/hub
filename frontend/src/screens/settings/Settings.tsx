import React from "react";
import { toast } from "sonner";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
import { StarsIcon } from "lucide-react";
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
import {
  DarkMode,
  Theme,
  Themes,
  useTheme,
} from "src/components/ui/theme-provider";
import {
  BITCOIN_DISPLAY_FORMAT_BIP177,
  BITCOIN_DISPLAY_FORMAT_SATS,
} from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useCurrencies } from "src/hooks/useCurrencies";
import { useInfo } from "src/hooks/useInfo";
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

  return (
    <>
      <SettingsHeader
        title="General"
        pageTitle="Settings"
        description="General Alby Hub settings."
      />
      <form className="w-full flex flex-col gap-8">
        {/* Theme & Appearance Section */}
        <div className="space-y-4">
          <h3 className="text-xl font-medium">Appearance</h3>
          <div className="space-y-4">
            <div className="grid gap-2">
              <Label htmlFor="theme">Theme</Label>
              <Select
                value={theme}
                onValueChange={(value) => {
                  if (paidThemes.includes(value) && !hasPlan) {
                    setShowUpgradeDialog(true);
                    return;
                  }
                  setTheme(value as Theme);
                  toast("Theme updated.");
                }}
              >
                <SelectTrigger className="w-full md:w-60">
                  <SelectValue placeholder="Theme" />
                </SelectTrigger>
                <SelectContent>
                  {Themes.map((theme) => {
                    const isPaidTheme = paidThemes.includes(theme);

                    return (
                      <SelectItem key={theme} value={theme}>
                        <div className="flex items-center justify-between gap-2 w-full">
                          <span className="capitalize">{theme}</span>
                          {isPaidTheme && (
                            <Badge variant="outline">
                              <StarsIcon />
                              Pro
                            </Badge>
                          )}
                        </div>
                      </SelectItem>
                    );
                  })}
                </SelectContent>
              </Select>
              <UpgradeDialog
                open={showUpgradeDialog}
                onOpenChange={setShowUpgradeDialog}
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="appearance">Appearance</Label>
              <Select
                value={darkMode}
                onValueChange={(value) => {
                  setDarkMode(value as DarkMode);
                  toast("Appearance updated.");
                }}
              >
                <SelectTrigger className="w-full md:w-60">
                  <SelectValue placeholder="Appearance" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="system">System</SelectItem>
                  <SelectItem value="light">Light</SelectItem>
                  <SelectItem value="dark">Dark</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        {/* Units & Currency Section */}
        <div className="space-y-4">
          <h3 className="text-xl font-medium">Units & Currency</h3>
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
      </form>
    </>
  );
}

export default Settings;
