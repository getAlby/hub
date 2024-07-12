import React, { useEffect } from "react";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { cn } from "src/lib/utils";
import {
  NIP_47_GET_BALANCE_METHOD,
  NIP_47_GET_INFO_METHOD,
  NIP_47_LIST_TRANSACTIONS_METHOD,
  NIP_47_LOOKUP_INVOICE_METHOD,
  NIP_47_MULTI_PAY_INVOICE_METHOD,
  NIP_47_MULTI_PAY_KEYSEND_METHOD,
  NIP_47_NOTIFICATIONS_PERMISSION,
  NIP_47_PAY_INVOICE_METHOD,
  NIP_47_PAY_KEYSEND_METHOD,
  ReadOnlyScope,
  SCOPE_GROUP_CUSTOM,
  SCOPE_GROUP_FULL_ACCESS,
  SCOPE_GROUP_READ_ONLY,
  Scope,
  ScopeGroupType,
  WalletCapabilities,
  scopeDescriptions,
  scopeGroupDescriptions,
  scopeGroupIconMap,
  scopeGroupTitle,
} from "src/types";

interface ScopesProps {
  capabilities: WalletCapabilities;
  scopes: Set<Scope>;
  onScopeChange: (scopes: Set<Scope>) => void;
}

const isSetEqual = (setA: Set<string>, setB: Set<string>) =>
  setA.size === setB.size && [...setA].every((value) => setB.has(value));

const Scopes: React.FC<ScopesProps> = ({
  capabilities,
  scopes,
  onScopeChange,
}) => {
  const fullAccessScopes: Set<Scope> = React.useMemo(() => {
    const scopes: Scope[] = capabilities.methods as Scope[];
    if (capabilities.notificationTypes.length) {
      scopes.push(NIP_47_NOTIFICATIONS_PERMISSION);
    }
    return new Set(
      scopes
        .map((scope) =>
          [
            NIP_47_PAY_KEYSEND_METHOD,
            NIP_47_MULTI_PAY_INVOICE_METHOD,
            NIP_47_MULTI_PAY_KEYSEND_METHOD,
          ].includes(scope)
            ? NIP_47_PAY_INVOICE_METHOD
            : scope
        )
        .filter((scope, i, self) => self.indexOf(scope) === i)
    );
  }, [capabilities]);

  const readOnlyScopes: Set<ReadOnlyScope> = React.useMemo(() => {
    const scopes: Scope[] = capabilities.methods as Scope[];
    if (capabilities.notificationTypes.length) {
      scopes.push(NIP_47_NOTIFICATIONS_PERMISSION);
    }
    return new Set(
      scopes.filter((method): method is ReadOnlyScope =>
        [
          NIP_47_GET_BALANCE_METHOD,
          NIP_47_GET_INFO_METHOD,
          NIP_47_LOOKUP_INVOICE_METHOD,
          NIP_47_LIST_TRANSACTIONS_METHOD,
          NIP_47_NOTIFICATIONS_PERMISSION,
        ].includes(method)
      )
    );
  }, [capabilities]);

  const [scopeGroup, setScopeGroup] = React.useState<ScopeGroupType>(() => {
    if (!scopes.size || isSetEqual(scopes, fullAccessScopes)) {
      return SCOPE_GROUP_FULL_ACCESS;
    } else if (isSetEqual(scopes, readOnlyScopes)) {
      return SCOPE_GROUP_READ_ONLY;
    }
    return SCOPE_GROUP_CUSTOM;
  });

  // we need scopes to be empty till this point for isScopesEditable
  // and once this component is mounted we set it to all scopes
  useEffect(() => {
    // stop setting scopes on re-renders
    if (!scopes.size && scopeGroup != SCOPE_GROUP_CUSTOM) {
      onScopeChange(fullAccessScopes);
    }
  }, [fullAccessScopes, onScopeChange, scopeGroup, scopes]);

  const handleScopeGroupChange = (scopeGroup: ScopeGroupType) => {
    setScopeGroup(scopeGroup);
    switch (scopeGroup) {
      case SCOPE_GROUP_FULL_ACCESS:
        onScopeChange(fullAccessScopes);
        break;
      case SCOPE_GROUP_READ_ONLY:
        onScopeChange(readOnlyScopes);
        break;
      default: {
        onScopeChange(new Set());
        break;
      }
    }
  };

  const handleScopeChange = (scope: Scope) => {
    const newScopes = new Set(scopes);
    if (newScopes.has(scope)) {
      newScopes.delete(scope);
    } else {
      newScopes.add(scope);
    }
    onScopeChange(newScopes);
  };

  return (
    <div className="mb-6">
      <div className="flex flex-col w-full mb-6">
        <p className="font-medium text-sm mb-2">Choose wallet permissions</p>
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {(
            [
              SCOPE_GROUP_FULL_ACCESS,
              SCOPE_GROUP_READ_ONLY,
              SCOPE_GROUP_CUSTOM,
            ] as ScopeGroupType[]
          ).map((sg, index) => {
            const ScopeGroupIcon = scopeGroupIconMap[sg];
            return (
              <div
                key={index}
                className={`flex flex-col items-center border-2 rounded cursor-pointer ${scopeGroup == sg ? "border-primary" : "border-muted"} p-4`}
                onClick={() => {
                  handleScopeGroupChange(sg);
                }}
              >
                <ScopeGroupIcon className="mb-2" />
                <p className="text-sm font-medium">{scopeGroupTitle[sg]}</p>
                <p className="text-[10px] text-muted-foreground text-nowrap">
                  {scopeGroupDescriptions[sg]}
                </p>
              </div>
            );
          })}
        </div>
      </div>

      {scopeGroup == "custom" && (
        <>
          <p className="font-medium text-sm">Authorize the app to:</p>
          <ul className="flex flex-col w-full mt-3">
            {capabilities.scopes.map((rm, index) => {
              return (
                <li
                  key={index}
                  className={cn(
                    "w-full",
                    rm == NIP_47_PAY_INVOICE_METHOD ? "order-last" : ""
                  )}
                >
                  <div className="flex items-center mb-2">
                    <Checkbox
                      id={rm}
                      className="mr-2"
                      onCheckedChange={() => handleScopeChange(rm)}
                      checked={scopes.has(rm)}
                    />
                    <Label htmlFor={rm} className="cursor-pointer">
                      {scopeDescriptions[rm]}
                    </Label>
                  </div>
                </li>
              );
            })}
            {capabilities.notificationTypes.length > 0 && (
              <li className="w-full">
                <div className="flex items-center mb-2">
                  <Checkbox
                    id={NIP_47_NOTIFICATIONS_PERMISSION}
                    className="mr-2"
                    onCheckedChange={() =>
                      handleScopeChange(NIP_47_NOTIFICATIONS_PERMISSION)
                    }
                    checked={scopes.has(NIP_47_NOTIFICATIONS_PERMISSION)}
                  />
                  <Label
                    htmlFor={NIP_47_NOTIFICATIONS_PERMISSION}
                    className="cursor-pointer"
                  >
                    {scopeDescriptions[NIP_47_NOTIFICATIONS_PERMISSION]}
                  </Label>
                </div>
              </li>
            )}
          </ul>
        </>
      )}
    </div>
  );
};

export default Scopes;
