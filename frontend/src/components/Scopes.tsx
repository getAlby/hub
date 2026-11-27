import {
  ArrowDownUpIcon,
  BrickWallIcon,
  ChevronsUpDownIcon,
  LucideIcon,
  MoveDownIcon,
  SquarePenIcon,
} from "lucide-react";
import React from "react";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "src/components/ui/sheet";
import { cn } from "src/lib/utils";
import {
  READ_ONLY_SCOPES,
  Scope,
  WalletCapabilities,
  scopeDescriptions,
  supportsHoldInvoices,
} from "src/types";

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
  isolated: "Isolated App",
  custom: "Custom",
};

const scopeGroupDescriptions: Record<ScopeGroup, string> = {
  full_access: "Send and receive payments",
  read_only: "Receive payments and view history",
  isolated: "Separate wallet with isolated balance",
  custom: "Define specific permissions",
};

type InvoiceScope = "make_invoice" | "make_invoice_only";

// hold invoice authority shown for a set of scopes (when the wallet supports hold invoices)
const holdInvoicesNote = (scopes: Scope[]) =>
  scopes.includes("make_invoice")
    ? "Includes hold invoices"
    : scopes.includes("make_invoice_only")
      ? "Hold invoices not included"
      : undefined;

interface ScopesProps {
  capabilities: WalletCapabilities;
  scopes: Scope[];
  isolated: boolean;
  onScopesChanged: (scopes: Scope[], isolated: boolean) => void;
}

const Scopes: React.FC<ScopesProps> = ({
  capabilities,
  scopes,
  isolated,
  onScopesChanged,
}) => {
  const [isSheetOpen, setSheetOpen] = React.useState(false);
  const holdInvoicesSupported = supportsHoldInvoices(capabilities);
  // Hold invoice authority (make_invoice) is never granted by ticking "Create
  // invoices" (make_invoice_only); it needs the separate hold invoices option.
  // Presets only keep it if the connection already had it when the editor was
  // opened and it was not removed since.
  const [presetInvoiceScope, setPresetInvoiceScope] =
    React.useState<InvoiceScope>(() =>
      scopes.includes("make_invoice") ? "make_invoice" : "make_invoice_only"
    );
  const withPresetInvoiceScope = React.useCallback(
    (presetScopes: Scope[]): Scope[] =>
      presetScopes.map((scope) =>
        scope === "make_invoice" ? presetInvoiceScope : scope
      ),
    [presetInvoiceScope]
  );

  const fullAccessScopes: Scope[] = React.useMemo(() => {
    return withPresetInvoiceScope([...capabilities.scopes]);
  }, [capabilities.scopes, withPresetInvoiceScope]);

  const readOnlyScopes: Scope[] = React.useMemo(() => {
    return withPresetInvoiceScope(
      capabilities.scopes.filter((scope) => READ_ONLY_SCOPES.includes(scope))
    );
  }, [capabilities.scopes, withPresetInvoiceScope]);

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

    return withPresetInvoiceScope(
      capabilities.scopes.filter((scope) => isolatedScopes.includes(scope))
    );
  }, [capabilities.scopes, withPresetInvoiceScope]);

  const presetScopes: Record<ScopeGroup, Scope[]> = {
    full_access: fullAccessScopes,
    read_only: readOnlyScopes,
    isolated: isolatedScopes,
    custom: [],
  };

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

  const isInvoiceScope = (scope: Scope) =>
    scope === "make_invoice" || scope === "make_invoice_only";
  const canCreateInvoices = scopes.some(isInvoiceScope);

  const handleScopeChange = (scope: Scope) => {
    let newScopes = [...scopes];
    if (scope === "make_invoice") {
      // "Create invoices": make_invoice_only; unticking also removes hold invoices
      if (canCreateInvoices) {
        newScopes = newScopes.filter((existing) => !isInvoiceScope(existing));
        setPresetInvoiceScope("make_invoice_only");
      } else {
        newScopes.push("make_invoice_only");
      }
    } else if (newScopes.includes(scope)) {
      newScopes = newScopes.filter((existing) => existing !== scope);
    } else {
      newScopes.push(scope);
    }

    onScopesChanged(newScopes, isolated);
  };

  const handleHoldInvoicesChange = () => {
    const invoiceScope: InvoiceScope = scopes.includes("make_invoice")
      ? "make_invoice_only"
      : "make_invoice";
    setPresetInvoiceScope(invoiceScope);
    onScopesChanged(
      [...scopes.filter((existing) => !isInvoiceScope(existing)), invoiceScope],
      isolated
    );
  };

  const ActiveScopeGroupIcon = scopeGroupIconMap[scopeGroup];
  const activeHoldInvoicesNote =
    holdInvoicesSupported && scopeGroup !== "custom"
      ? holdInvoicesNote(scopes)
      : undefined;

  return (
    <>
      <Sheet open={isSheetOpen} onOpenChange={setSheetOpen}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>Choose wallet permissions</SheetTitle>
          </SheetHeader>
          <div className="flex flex-col gap-4 px-6">
            {scopeGroups.map((sg, index) => {
              const ScopeGroupIcon = scopeGroupIconMap[sg];
              const holdNote = holdInvoicesSupported
                ? holdInvoicesNote(presetScopes[sg])
                : undefined;
              return (
                <button
                  type="button"
                  key={index}
                  className={cn(
                    "flex gap-3 items-center rounded-lg border-2 cursor-pointer p-4",
                    scopeGroup == sg ? "border-primary" : "border-muted"
                  )}
                  onClick={() => {
                    handleScopeGroupChange(sg);
                    setSheetOpen(false);
                  }}
                >
                  <ScopeGroupIcon className="shrink-0 size-5 text-muted-foreground" />
                  <div className="flex flex-col gap-0.5 text-left">
                    <p className="text-sm font-medium">{scopeGroupTitle[sg]}</p>
                    <span className="text-xs text-muted-foreground">
                      {scopeGroupDescriptions[sg]}
                    </span>
                    {holdNote && (
                      <span className="text-xs text-muted-foreground">
                        {holdNote}
                      </span>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
          <SheetFooter>
            <Button type="submit">Save changes</Button>
            <SheetClose asChild>
              <Button variant="outline">Close</Button>
            </SheetClose>
          </SheetFooter>
        </SheetContent>
      </Sheet>
      <button
        type="button"
        className="flex items-center gap-2 rounded-lg border cursor-pointer p-4 w-full"
        onClick={() => {
          setSheetOpen(true);
        }}
      >
        <ActiveScopeGroupIcon className="shrink-0 size-5 text-muted-foreground" />
        <div className="flex flex-col gap-0.5 text-left">
          <p className="text-sm font-medium">{scopeGroupTitle[scopeGroup]}</p>
          <span className="text-xs text-muted-foreground">
            {scopeGroupDescriptions[scopeGroup]}
          </span>
          {activeHoldInvoicesNote && (
            <span className="text-xs text-muted-foreground">
              {activeHoldInvoicesNote}
            </span>
          )}
        </div>
        <ChevronsUpDownIcon className="ml-auto shrink-0 size-4 text-muted-foreground" />
      </button>

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
                      checked={
                        scope === "make_invoice"
                          ? canCreateInvoices
                          : scopes.includes(scope)
                      }
                    />
                    <Label htmlFor={scope} className="cursor-pointer">
                      {scope === "make_invoice"
                        ? scopeDescriptions.make_invoice_only
                        : scopeDescriptions[scope]}
                    </Label>
                  </div>
                  {scope === "make_invoice" &&
                    holdInvoicesSupported &&
                    canCreateInvoices && (
                      <div className="flex items-center mb-2 ml-6">
                        <Checkbox
                          id="hold_invoices"
                          className="mr-2"
                          onCheckedChange={handleHoldInvoicesChange}
                          checked={scopes.includes("make_invoice")}
                        />
                        <Label
                          htmlFor="hold_invoices"
                          className="cursor-pointer"
                        >
                          Also allow hold invoices
                        </Label>
                      </div>
                    )}
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
