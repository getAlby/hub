import { type Theme, useTheme } from "src/components/ui/theme-provider";
import { cn } from "src/lib/utils";

export function ThemePreview({ theme }: { theme: Theme }) {
  const { isDarkMode } = useTheme();
  const darkClass = isDarkMode ? "dark" : "";

  return (
    <div
      aria-hidden="true"
      className={cn(
        `theme-${theme}`,
        darkClass,
        "w-full aspect-16/10 flex bg-background"
      )}
    >
      <div className="w-1/4 flex flex-col p-1 gap-0.5 bg-sidebar border-r border-border">
        <div className="h-1 w-3/4 rounded-full bg-sidebar-primary" />
        <div className="h-0.5 w-1/2 rounded-full mt-1 bg-muted" />
        <div className="h-0.5 w-2/3 rounded-full bg-muted" />
        <div className="h-0.5 w-1/2 rounded-full bg-muted" />
      </div>
      <div className="flex-1 p-1.5 flex flex-col gap-1">
        <div className="h-1 w-1/2 rounded-full bg-foreground/20" />
        <div className="mt-0.5 rounded flex-1 p-1 flex flex-col gap-0.5 bg-card border border-border">
          <div className="h-0.5 w-3/4 rounded-full bg-muted" />
          <div className="h-0.5 w-1/2 rounded-full bg-muted" />
          <div className="mt-auto h-1 w-8 rounded-sm bg-primary" />
        </div>
      </div>
    </div>
  );
}
