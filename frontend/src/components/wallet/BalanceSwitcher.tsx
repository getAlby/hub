import { BitcoinIcon, ZapIcon } from "lucide-react";
import { Link } from "react-router";
import { cn } from "src/lib/utils";

type Mode = "lightning" | "onchain";

const items: { mode: Mode; to: string; label: string; Icon: typeof ZapIcon }[] =
  [
    { mode: "lightning", to: "/wallet", label: "Lightning", Icon: ZapIcon },
    {
      mode: "onchain",
      to: "/wallet/onchain",
      label: "On-chain",
      Icon: BitcoinIcon,
    },
  ];

export function BalanceSwitcher({ active }: { active: Mode }) {
  return (
    <div className="inline-flex items-center gap-1 rounded-lg bg-muted p-0.75">
      {items.map(({ mode, to, label, Icon }) => {
        const isActive = mode === active;
        return (
          <Link
            key={mode}
            to={to}
            aria-current={isActive ? "page" : undefined}
            className={cn(
              "inline-flex h-7 items-center gap-1.5 rounded-md px-3 text-xs font-medium transition-colors",
              isActive
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <Icon className="size-3.5" />
            {label}
          </Link>
        );
      })}
    </div>
  );
}
