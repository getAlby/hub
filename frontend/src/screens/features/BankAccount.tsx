import { CreditCardIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import EmptyState from "src/components/EmptyState";

function BankAccount() {
  return (
    <>
      <AppHeader
        title="Virtual Bank Account 🌐"
        description="Bridge Alby Hub to legacy finance using a virtual bank account and virtual credit cards."
      />

      <EmptyState
        icon={CreditCardIcon}
        title="Activate Your Virtual Bank Account"
        description="Fund your Hub with fiat, settle traditional payments, and move money to Lightning in seconds. Virtual credit card coming soon!"
        buttonText="Sign up for waitlist"
        buttonLink="/connect-bank"
      />
    </>
  );
}

export default BankAccount;
