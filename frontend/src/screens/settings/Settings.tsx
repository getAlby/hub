import { useTranslation } from "react-i18next";
import LocaleSwitcher from "src/components/LocaleSwitcher";
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
import { toast } from "src/components/ui/use-toast";

function Settings() {
  const { t } = useTranslation("translation", { keyPrefix: "settings" });
  const { theme, darkMode, setTheme, setDarkMode } = useTheme();

  return (
    <>
      <SettingsHeader
        title={t("theme.title")}
        description={t("theme.description")}
      />
      <form className="w-full flex flex-col gap-3 mb-4">
        <div className="grid gap-1.5">
          <Label htmlFor="theme">{t("theme.title")}</Label>
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
              <SelectItem value="system">
                {t("theme.options.system")}
              </SelectItem>
              <SelectItem value="light">{t("theme.options.light")}</SelectItem>
              <SelectItem value="dark">{t("theme.options.dark")}</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </form>
      <SettingsHeader
        title={t("language.title")}
        description={t("language.description")}
      />
      <form className="w-full flex flex-col gap-3">
        <div className="grid gap-1.5">
          <Label htmlFor="language">{t("language.title")}</Label>
          <LocaleSwitcher />
        </div>
      </form>
    </>
  );
}

export default Settings;
