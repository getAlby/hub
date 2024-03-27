import { Switch } from "src/components/ui/switch";
import { useTheme } from "src/components/ui/theme-provider";
import { Label } from "src/components/ui/label";

export function ModeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <>
      <div className="flex items-center space-x-2 my-5 mx-3">
        <Switch
          id="dark-mode"
          onClick={() => {
            setTheme(
              theme === "system" ? "dark" : theme == "dark" ? "light" : "dark"
            );
          }}
        />
        <Label htmlFor="dark-mode" className=" cursor-pointer">
          Dark Mode
        </Label>
      </div>
    </>
  );
}
