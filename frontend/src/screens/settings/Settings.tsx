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
import { request } from "src/utils/request";

function Settings() {
  const { theme, darkMode, setTheme, setDarkMode } = useTheme();
  const { toast } = useToast();
  const [currencies, setCurrencies] = useState<[string, string][]>([]);
  const [filteredCurrencies, setFilteredCurrencies] = useState<
    [string, string][]
  >([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const { data: info } = useInfo();

  const [selectedCurrency, setSelectedCurrency] = useState<string | undefined>(
    info?.currency
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
        const data = await response.json();

        const mappedCurrencies: [string, string][] = Object.entries(data).map(
          ([code, details]: any) => [code.toUpperCase(), details.name]
        );

        mappedCurrencies.sort((a, b) => a[1].localeCompare(b[1]));

        setCurrencies(mappedCurrencies);
        setFilteredCurrencies(mappedCurrencies);
      } catch (error) {
        console.error(error || "Failed to fetch currencies");
      } finally {
        setLoading(false);
      }
    }

    fetchCurrencies();
  }, []);

  async function updateCurrency(currency: string) {
    try {
      const response = await request("/api/currency", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ currency }),
      });
    } catch (error) {
      console.error(error);
      throw error;
    }
  }

  useEffect(() => {
    const filtered = currencies.filter(
      ([code, name]) =>
        name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        code.toLowerCase().includes(searchQuery.toLowerCase())
    );
    setFilteredCurrencies(filtered);
  }, [searchQuery, currencies]);

  console.log(filteredCurrencies);

  return (
    <>
      <SettingsHeader
        title="Theme"
        description="Alby Hub is your wallet, make it your style."
      />
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
          <Label htmlFor="currency">Select Currency</Label>
          {loading ? (
            <p>Loading currencies...</p>
          ) : (
            <Select
              value={selectedCurrency}
              onValueChange={async (value) => {
                setSelectedCurrency(value);
                console.log(value);
                await updateCurrency(value);
                toast({ title: `Currency set to ${value}` });
              }}
            >
              <SelectTrigger className="w-[250px] border border-gray-300 p-2 rounded-md">
                <SelectValue placeholder="Select a currency" />
              </SelectTrigger>
              <SelectContent>
                {filteredCurrencies.map(([code, name]) => (
                  <SelectItem key={code} value={code}>
                    {name} ({code})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>
      </form>
    </>
  );
}

export default Settings;
