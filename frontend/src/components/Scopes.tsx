import React from "react";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { cn } from "src/lib/utils";
import {
  NIP_47_MAKE_INVOICE_METHOD,
  NIP_47_NOTIFICATIONS_PERMISSION,
  NIP_47_PAY_INVOICE_METHOD,
  SCOPE_GROUP_CUSTOM,
  SCOPE_GROUP_ONLY_RECEIVE,
  SCOPE_GROUP_SEND_RECEIVE,
  Scope,
  ScopeGroupType,
  WalletCapabilities,
  scopeDescriptions,
  scopeGroupDescriptions,
  scopeGroupIconMap,
  scopeGroupTitle,
} from "src/types";

// TODO: this runs everytime, use useEffect
const scopeGrouper = (scopes: Set<Scope>) => {
  if (
    scopes.size === 2 &&
    scopes.has(NIP_47_MAKE_INVOICE_METHOD) &&
    scopes.has(NIP_47_PAY_INVOICE_METHOD)
  ) {
    return "send_receive";
  } else if (scopes.size === 1 && scopes.has(NIP_47_MAKE_INVOICE_METHOD)) {
    return "only_receive";
  }
  return "custom";
};

const validScopeGroups = (capabilities: WalletCapabilities) => {
  const scopeGroups = [SCOPE_GROUP_CUSTOM];
  if (capabilities.scopes.includes(NIP_47_MAKE_INVOICE_METHOD)) {
    scopeGroups.unshift(SCOPE_GROUP_ONLY_RECEIVE);
    if (capabilities.scopes.includes(NIP_47_PAY_INVOICE_METHOD)) {
      scopeGroups.unshift(SCOPE_GROUP_SEND_RECEIVE);
    }
  }
  return scopeGroups;
};

interface ScopesProps {
  capabilities: WalletCapabilities;
  scopes: Set<Scope>;
  onScopeChange: (scopes: Set<Scope>) => void;
}

const Scopes: React.FC<ScopesProps> = ({
  capabilities,
  scopes,
  onScopeChange,
}) => {
  const [scopeGroup, setScopeGroup] = React.useState(scopeGrouper(scopes));
  const scopeGroups = validScopeGroups(capabilities);

  // TODO: EDITABLE PROP
  const handleScopeGroupChange = (scopeGroup: ScopeGroupType) => {
    setScopeGroup(scopeGroup);
    switch (scopeGroup) {
      case "send_receive":
        onScopeChange(
          new Set([NIP_47_PAY_INVOICE_METHOD, NIP_47_MAKE_INVOICE_METHOD])
        );
        break;
      case "only_receive":
        onScopeChange(new Set([NIP_47_MAKE_INVOICE_METHOD]));
        break;
      default: {
        const newSet = new Set(capabilities.scopes);
        if (capabilities.notificationTypes.length) {
          newSet.add(NIP_47_NOTIFICATIONS_PERMISSION);
        }
        onScopeChange(newSet);
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
      {scopeGroups.length > 1 && (
        <div className="flex flex-col w-full mb-6">
          <p className="font-medium text-sm mb-2">Choose wallet permissions</p>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            {(scopeGroups as ScopeGroupType[]).map((sg, index) => {
              if (
                scopeGroup == SCOPE_GROUP_SEND_RECEIVE &&
                !capabilities.scopes.includes(NIP_47_PAY_INVOICE_METHOD)
              ) {
                return;
              }
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
      )}

      {(scopeGroup == "custom" || scopeGroups.length == 1) && (
        <>
          <p className="font-medium text-sm">Authorize the app to:</p>
          <ul className="flex flex-col w-full mt-3">
            {capabilities.scopes.map((rm, index) => {
              return (
                <li
                  key={index}
                  className={cn(
                    "w-full",
                    rm == "pay_invoice" ? "order-last" : ""
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
