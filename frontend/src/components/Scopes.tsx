import {
  ArrowDownUpIcon,
  BrickWallIcon,
  LucideIcon,
  MoveDownIcon,
  SquarePenIcon,
} from "lucide-react";
import React from "react";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { cn } from "src/lib/utils";
import { Scope, WalletCapabilities, scopeDescriptions } from "src/types";

const scopeGroups = ["full_access", "read_only", "isolated", "custom"] as const;
type ScopeGroup = (typeof scopeGroups)[number];
type ScopeGroupIconMap = { [key in ScopeGroup]: LucideIcon };

const scopeGroupIconMap: ScopeGroupIconMap = {
  full_access: ArrowDownUpIcon,
  read_only: MoveDownIcon,
  isolated: BrickWallIcon,
  custom: SquarePenIcon,
};

const scopeGroupTitle: Record<ScopeGroup, string> = {
  full_access: "Full Access",
  read_only: "Read Only",
  isolated: "Isolated",
  custom: "Custom",
};

const scopeGroupDescriptions: Record<ScopeGroup, string> = {
  full_access: "I trust this app to access my wallet within the budget I set",
  read_only: "This app can receive payments and read my transaction history",
  isolated:
    "This app will have its own balance and only sees its own transactions",
  custom: "I want to define exactly what access this app has to my wallet",
};

interface ScopesProps {
  capabilities: WalletCapabilities;
  scopes: Scope[];
  isolated: boolean;
  isNewConnection: boolean;
  onScopesChanged: (scopes: Scope[], isolated: boolean) => void;
}

const Scopes: React.FC<ScopesProps> = ({
  capabilities,
  scopes,
  isolated,
  onScopesChanged,
}) => {
  const fullAccessScopes: Scope[] = React.useMemo(() => {
    return [...capabilities.scopes];
  }, [capabilities.scopes]);

  const readOnlyScopes: Scope[] = React.useMemo(() => {
    const readOnlyScopes: Scope[] = [
      "get_balance",
      "get_info",
      "make_invoice",
      "lookup_invoice",
      "list_transactions",
      "notifications",
    ];

    return capabilities.scopes.filter((scope) =>
      readOnlyScopes.includes(scope)
    );
  }, [capabilities.scopes]);

  const isolatedScopes: Scope[] = React.useMemo(() => {
    const isolatedScopes: Scope[] = [
      "pay_invoice",
      "get_balance",
      "get_info",
      "make_invoice",
      "lookup_invoice",
      "list_transactions",
      "notifications",
    ];

    return capabilities.scopes.filter((scope) =>
      isolatedScopes.includes(scope)
    );
  }, [capabilities.scopes]);

  const [scopeGroup, setScopeGroup] = React.useState<ScopeGroup>(() => {
    if (
      isolated &&
      scopes.length === isolatedScopes.length &&
      scopes.every((scope) => isolatedScopes.includes(scope))
    ) {
      return "isolated";
    }
    if (
      scopes.length === fullAccessScopes.length &&
      scopes.every((scope) => fullAccessScopes.includes(scope))
    ) {
      return "full_access";
    }
    if (
      scopes.length === readOnlyScopes.length &&
      readOnlyScopes.every((readOnlyScope) => scopes.includes(readOnlyScope))
    ) {
      return "read_only";
    }

    return "custom";
  });

  const handleScopeGroupChange = (scopeGroup: ScopeGroup) => {
    setScopeGroup(scopeGroup);
    switch (scopeGroup) {
      case "full_access":
        onScopesChanged(fullAccessScopes, false);
        break;
      case "read_only":
        onScopesChanged(readOnlyScopes, false);
        break;
      case "isolated":
        onScopesChanged(isolatedScopes, true);
        break;
      default: {
        onScopesChanged([], false);
        break;
      }
    }
  };

  const handleScopeChange = (scope: Scope) => {
    let newScopes = [...scopes];
    if (newScopes.includes(scope)) {
      newScopes = newScopes.filter((existing) => existing !== scope);
    } else {
      newScopes.push(scope);
    }

    onScopesChanged(newScopes, isolated);
  };

  return (
    <>
      <div className="flex flex-col w-full mb-4">
        <p className="font-medium text-sm mb-2">Choose wallet permissions</p>
        <div className="grid grid-cols-2 gap-4">
          {scopeGroups.map((sg, index) => {
            const ScopeGroupIcon = scopeGroupIconMap[sg];
            return (
              <button
                key={index}
                className={`flex flex-col items-center border-2 rounded cursor-pointer ${scopeGroup == sg ? "border-primary" : "border-muted"} p-4`}
                onClick={() => {
                  handleScopeGroupChange(sg);
                }}
              >
                <ScopeGroupIcon className="mb-2" />
                <p className="text-sm font-medium">{scopeGroupTitle[sg]}</p>
                <span className="w-full text-center text-xs text-muted-foreground">
                  {scopeGroupDescriptions[sg]}
                </span>
              </button>
            );
          })}
        </div>
      </div>

      {scopeGroup == "custom" && (
        <div className="mb-2">
          <p className="font-medium text-sm mt-4">Isolation</p>
          <div className="flex items-center mt-2">
            <Checkbox
              id="isolated"
              className="mr-2"
              onCheckedChange={() => onScopesChanged(scopes, !isolated)}
              checked={isolated}
            />
            <Label htmlFor="isolated" className="cursor-pointer">
              Isolate this app's balance and transactions
            </Label>
          </div>
          <p className="font-medium text-sm mt-4">Authorize the app to:</p>
          <ul className="flex flex-col w-full mt-2">
            {capabilities.scopes.map((scope, index) => {
              return (
                <li
                  key={index}
                  className={cn(
                    "w-full",
                    scope == "pay_invoice" ? "order-last" : ""
                  )}
                >
                  <div className="flex items-center mb-2">
                    <Checkbox
                      id={scope}
                      className="mr-2"
                      onCheckedChange={() => handleScopeChange(scope)}
                      checked={scopes.includes(scope)}
                    />
                    <Label htmlFor={scope} className="cursor-pointer">
                      {scopeDescriptions[scope]}
                    </Label>
                  </div>
                </li>
              );
            })}
          </ul>
        </div>
      )}
    </>
  );
};

export default Scopes;
