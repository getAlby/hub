import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";
import { getSwapProviderName } from "src/utils/getSwapProviderName";

export function SwapProviderAttribution({ className }: { className?: string }) {
  const { data: info } = useInfo();
  if (!info?.swapApiUrl) {
    return null;
  }

  return (
    <p className={cn("text-xs text-muted-foreground", className)}>
      powered by{" "}
      <span className="font-medium text-foreground">
        {getSwapProviderName(info.swapApiUrl)}
      </span>
    </p>
  );
}
