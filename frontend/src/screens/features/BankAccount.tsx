import { CreditCardIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import { request } from "src/utils/request";

function BankAccount() {
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
  }

  return (
    <>
      <AppHeader
        title="Virtual Bank Account 🌐"
        description="Bridge Alby Hub to legacy finance using a virtual bank account and virtual credit cards."
      />
      <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm p-8">
        <div className="flex flex-col items-center gap-1 text-center max-w-sm">
          <CreditCardIcon className="w-10 h-10 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">
            Activate Your Virtual Bank Account
          </h3>
          <p className="text-sm text-muted-foreground">
            Fund your Hub with fiat, settle traditional payments, and move money
            to Lightning in seconds. Virtual credit card coming soon!
          </p>
          <Button onClick={activate} className="mt-4">
            Request Early Access
          </Button>
        </div>
      </div>
    </>
  );
}

export default BankAccount;
