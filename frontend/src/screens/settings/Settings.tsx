import SettingsHeader from "src/components/SettingsHeader";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useTheme } from "src/components/ui/theme-provider";

function Settings() {
  const { theme, setTheme } = useTheme();

  return (
    <>
      <SettingsHeader
        title="Theme"
        description="Alby Hub is your wallet, make it your style."
      />
      <form className="w-full flex flex-col gap-3">
        <div className="grid gap-1.5">
          <Label htmlFor="theme">Theme</Label>
          <Select value={theme} onValueChange={(value) => setTheme(value)}>
            <SelectTrigger className="w-[150px]">
              <SelectValue placeholder="Theme" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem key="system" value="system">
                System
              </SelectItem>
              <SelectItem key="dark" value="dark">
                Dark
              </SelectItem>
              <SelectItem key="light" value="light">
                Light
              </SelectItem>
              <SelectItem key="theme-nostr-light" value="theme-nostr-light">
                Nostr (light)
              </SelectItem>
              <SelectItem key="theme-nostr-dark" value="theme-nostr-dark">
                Nostr (dark)
              </SelectItem>
              <SelectItem key="theme-dracula-light" value="theme-dracula-light">
                Dracula (light)
              </SelectItem>
              <SelectItem key="theme-dracula-dark" value="theme-dracula-dark">
                Dracula (dark)
              </SelectItem>
              <SelectItem key="theme-nord-light" value="theme-nord-light">
                Nord (light)
              </SelectItem>
              <SelectItem key="theme-nord-dark" value="theme-nord-dark">
                Nord (dark)
              </SelectItem>
              <SelectItem key="theme-purple" value="theme-purple">
                Purple
              </SelectItem>
              <SelectItem key="theme-blue" value="theme-blue">
                Blue
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </form>
    </>
  );
}

export default Settings;
