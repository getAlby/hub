import { BitcoinIcon, ZapIcon } from "lucide-react";
import { Link } from "react-router";
import {
  useSwitcherVariant,
  type Variant,
} from "src/components/wallet/useSwitcherVariant";
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

/**
 * Option A: icon-only segmented switcher.
 * Replaces the "Spending Balance ⇅" / "On-chain Balance ⇅" toggle.
 */
export function IconBalanceSwitcher({ active }: { active: Mode }) {
  return (
    <div className="inline-flex items-center gap-1 rounded-full border bg-muted/40 p-1">
      {items.map(({ mode, to, label, Icon }) => {
        const isActive = mode === active;
        return (
          <Link
            key={mode}
            to={to}
            aria-label={label}
            aria-current={isActive ? "page" : undefined}
            className={cn(
              "inline-flex size-8 items-center justify-center rounded-full transition-colors",
              isActive
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground"
            )}
          >
            <Icon className="size-4" />
          </Link>
        );
      })}
    </div>
  );
}

/**
 * Option B: tab-style switcher with icon + label.
 * Sits in the same slot as the icon-only version but reads as navigation.
 */
export function TabBalanceSwitcher({ active }: { active: Mode }) {
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

/**
 * Floating dev toggle to preview the three switcher variants.
 * Remove before merging.
 */
export function SwitcherVariantToggle() {
  const [variant, setVariant] = useSwitcherVariant();
  const variants: Variant[] = ["icons", "tabs", "original"];
  return (
    <div className="fixed bottom-4 right-4 z-50 flex gap-1 rounded-md border bg-background/95 p-1 text-xs shadow-md backdrop-blur">
      {variants.map((v) => (
        <button
          key={v}
          type="button"
          onClick={() => setVariant(v)}
          className={cn(
            "rounded px-2 py-1 capitalize transition-colors",
            variant === v
              ? "bg-primary text-primary-foreground"
              : "text-muted-foreground hover:text-foreground"
          )}
        >
          {v}
        </button>
      ))}
    </div>
  );
}
