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
  const { theme, darkMode, setTheme, setDarkMode } = useTheme();

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
      </form>
    </>
  );
}

export default Settings;
