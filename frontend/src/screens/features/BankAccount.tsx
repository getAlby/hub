import { CheckCircle, CreditCardIcon } from "lucide-react";
import { useState } from "react";
import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import { localStorageKeys } from "src/constants";
import { request } from "src/utils/request";

function BankAccount() {
  const [activated, setActivated] = useState(
    !!localStorage.getItem(localStorageKeys.interestVirtualBankAccount)
  );

  async function activate() {
    await request(`/api/event`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        event: "interest_virtual_bankaccount",
      }),
    });

    localStorage.setItem(localStorageKeys.interestVirtualBankAccount, "true");
    setActivated(true);
  }

  return (
    <>
      <AppHeader
        title="Bank Account"
        description="Receive money from your bank account, pay bills and send payments to bank accounts. Buy and sell bitcoin at competitive rates."
      />
      <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-xs p-8">
        <div className="flex flex-col items-center gap-1 text-center max-w-sm">
          {!activated ? (
            <>
              <CreditCardIcon className="w-10 h-10 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">
                Activate Your Virtual Bank Account
              </h3>
              <p className="text-sm text-muted-foreground">
                Fund your Hub with fiat, settle traditional payments and buy or
                sell bitcoin at competitive rates!
              </p>
              <Button onClick={activate} className="mt-4">
                Request Early Access
              </Button>
            </>
          ) : (
            <>
              <CheckCircle className="w-10 h-10 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-semibold">
                Thanks for your interest!
              </h3>
              <p className="text-sm text-muted-foreground">
                We'll let you know as soon as this feature is available.
              </p>
            </>
          )}
        </div>
      </div>
    </>
  );
}

export default BankAccount;
