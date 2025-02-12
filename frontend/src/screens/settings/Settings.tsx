import { useEffect, useState } from "react";
import SettingsHeader from "src/components/SettingsHeader";
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
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function Settings() {
  const { theme, darkMode, setTheme, setDarkMode } = useTheme();
  const { toast } = useToast();

  const [fiatCurrencies, setFiatCurrencies] = useState<[string, string][]>([]);

  const { data: info } = useInfo();

  const [selectedCurrency, setSelectedCurrency] = useState<string | undefined>(
    info?.currency.toUpperCase()
  );

  useEffect(() => {
    if (info?.currency) {
      setSelectedCurrency(info.currency.toUpperCase());
    }
  }, [info]);

  useEffect(() => {
    async function fetchCurrencies() {
      try {
        const response = await fetch(`https://getalby.com/api/rates`);
        const data: Record<string, { name: string }> = await response.json();

        const mappedCurrencies: [string, string][] = Object.entries(data).map(
          ([code, details]) => [code.toUpperCase(), details.name]
        );

        mappedCurrencies.sort((a, b) => a[1].localeCompare(b[1]));

        setFiatCurrencies(mappedCurrencies);
      } catch (error) {
        console.error(error);
        handleRequestError(toast, "Failed to fetch currencies", error);
      }
    }

    fetchCurrencies();
  }, [toast]);

  async function updateCurrency(currency: string) {
    try {
      await request("/api/settings", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ currency }),
      });
    } catch (error) {
      console.error(error);
      handleRequestError(toast, "Failed to update currencies", error);
    }
  }

  return (
    <>
      <SettingsHeader title="General" description="General Alby Hub Settings" />
      <form className="w-full flex flex-col gap-3">
        <div className="grid gap-1.5">
          <Label htmlFor="theme">Theme</Label>
          <Select
            value={theme}
            onValueChange={(value) => {
              setTheme(value as Theme);
              toast({ title: "Theme updated." });
            }}
          >
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Theme" />
            </SelectTrigger>
            <SelectContent>
              {Themes.map((theme) => (
                <SelectItem key={theme} value={theme}>
                  {theme.charAt(0).toUpperCase() + theme.substring(1)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="theme">Dark mode</Label>
          <Select
            value={darkMode}
            onValueChange={(value) => {
              setDarkMode(value as DarkMode);
              toast({ title: "Dark Mode updated." });
            }}
          >
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Dark mode" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="system">System</SelectItem>
              <SelectItem value="light">Light</SelectItem>
              <SelectItem value="dark">Dark</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="currency">Fiat Currency</Label>
          <Select
            value={selectedCurrency}
            onValueChange={async (value) => {
              setSelectedCurrency(value);
              await updateCurrency(value);
              toast({ title: `Currency set to ${value}` });
            }}
          >
            <SelectTrigger className="w-[250px] border border-gray-300 p-2 rounded-md">
              <SelectValue placeholder="Select a currency" />
            </SelectTrigger>
            <SelectContent>
              {fiatCurrencies.map(([code, name]) => (
                <SelectItem key={code} value={code}>
                  {name} ({code})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </form>
    </>
  );
}

export default Settings;
