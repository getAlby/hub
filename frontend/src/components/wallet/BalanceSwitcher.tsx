import { BitcoinIcon, type LucideIcon, ZapIcon } from "lucide-react";
import { Link } from "react-router";
import { Tabs, TabsList, TabsTrigger } from "src/components/ui/tabs";

type Mode = "lightning" | "onchain";

const items: { mode: Mode; to: string; label: string; Icon: LucideIcon }[] = [
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
    <Tabs value={active}>
      <TabsList>
        {items.map(({ mode, to, label, Icon }) => (
          <TabsTrigger key={mode} value={mode} asChild>
            <Link to={to}>
              <Icon />
              {label}
            </Link>
          </TabsTrigger>
        ))}
      </TabsList>
    </Tabs>
  );
}
