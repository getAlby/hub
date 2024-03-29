import { Moon, Sun } from "lucide-react";
import { Switch } from "src/components/ui/switch";
import { useTheme } from "src/components/ui/theme-provider";

export function ModeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <>
      <div className="flex items-center space-x-2 text-muted-foreground">
        <Moon className="w-4 h-4" />
        <Switch
          onClick={() => {
            setTheme(
              theme === "system" ? "dark" : theme === "dark" ? "light" : "dark"
            );
          }}
        />
        <Sun className="w-4 h-4" />
      </div>
    </>
  );
}
