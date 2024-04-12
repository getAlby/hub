import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import { LifeBuoy, ShieldCheck } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import MnemonicInputs from "src/components/MnemonicInputs";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Alert } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { useToast } from "src/components/ui/use-toast";
import useSetupStore from "src/state/SetupStore";

export function ImportMnemonic() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const setupStore = useSetupStore();

  const [mnemonic, setMnemonic] = useState<string>("");

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (
      mnemonic.split(" ").length !== 12 ||
      !bip39.validateMnemonic(mnemonic, wordlist)
    ) {
      toast({
        title: "Invalid recovery phrase",
        variant: "destructive",
      });
      return;
    }

    const currentDate = new Date();
    const sixMonthsLater = new Date(
      currentDate.setMonth(currentDate.getMonth() + 6)
    );

    setupStore.updateNodeInfo({
      mnemonic,
      nextBackupReminder: sixMonthsLater.toISOString(),
    });
    navigate(`/setup/finish`);
  }

  return (
    <>
      <form
        onSubmit={onSubmit}
        className="flex flex-col gap-5 mx-auto max-w-2xl text-sm"
      >
        <TwoColumnLayoutHeader
          title="Import Master Key"
          description="Enter the your Master Key recovery phrase to import your Alby Hub."
        />

        <Alert>
          <div className="flex flex-col gap-4">
            <div className="flex gap-2 items-center">
              <div className="shrink-0 text-muted-foreground">
                <LifeBuoy className="w-6 h-6" />
              </div>
              <span className="text-muted-foreground">
                Recovery phrase is a set of 12 words that{" "}
                <b>restores your wallet from a backup</b>
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <div className="shrink-0 text-muted-foreground">
                <ShieldCheck className="w-6 h-6" />
              </div>
              <span className="text-muted-foreground">
                Make sure to enter them somewhere safe and private
              </span>
            </div>
          </div>
        </Alert>

        <MnemonicInputs mnemonic={mnemonic} setMnemonic={setMnemonic} />
        <Button>Finish</Button>
      </form>
    </>
  );
}
