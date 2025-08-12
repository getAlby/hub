import { StarsIcon } from "lucide-react";
import { useEffect, useState } from "react";
import Loading from "src/components/Loading";
import SettingsHeader from "src/components/SettingsHeader";
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
import { useToast } from "src/components/ui/use-toast";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

function Settings() {
  const { data: albyMe } = useAlbyMe();
  const { theme, darkMode, setTheme, setDarkMode } = useTheme();
  const { toast } = useToast();

  const [fiatCurrencies, setFiatCurrencies] = useState<[string, string][]>([]);

  const { data: info, mutate: reloadInfo } = useInfo();

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
      await reloadInfo();
      toast({ title: `Currency set to ${currency}` });
    } catch (error) {
      console.error(error);
      handleRequestError(toast, "Failed to update currencies", error);
    }
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
        description="General Alby Hub settings."
      />
      <form className="w-full flex flex-col gap-4">
        <div className="grid gap-2">
          <Label htmlFor="theme">Theme</Label>
          <Select
            value={theme}
            onValueChange={(value) => {
              setTheme(value as Theme);
              toast({ title: "Theme updated." });
            }}
          >
            <SelectTrigger className="w-full md:w-60">
              <SelectValue placeholder="Theme" />
            </SelectTrigger>
            <SelectContent>
              {Themes.map((theme) => {
                const isPaidTheme = paidThemes.includes(theme);
                const isDisabled = isPaidTheme && !hasPlan;

                return (
                  <SelectItem key={theme} value={theme} disabled={isDisabled}>
                    <div className="flex items-center justify-between gap-2 w-full">
                      <span
                        className={isDisabled ? "text-muted-foreground" : ""}
                      >
                        {theme.charAt(0).toUpperCase() + theme.substring(1)}
                      </span>
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
        </div>
        <div className="grid gap-2">
          <Label htmlFor="theme">Appearance</Label>
          <Select
            value={darkMode}
            onValueChange={(value) => {
              setDarkMode(value as DarkMode);
              toast({ title: "Appearance updated." });
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
        <div className="grid gap-1.5">
          <Label htmlFor="currency">Fiat Currency</Label>
          <Select value={info?.currency} onValueChange={updateCurrency}>
            <SelectTrigger className="w-full md:w-60">
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
