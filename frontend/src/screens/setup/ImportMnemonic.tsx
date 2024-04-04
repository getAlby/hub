import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import { LifeBuoy, ShieldCheck } from "lucide-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import MnemonicInputs from "src/components/MnemonicInputs";
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
        className="flex flex-col gap-2 mx-auto max-w-2xl text-sm"
      >
        <h1 className="font-semibold text-2xl font-headline mb-2 dark:text-white">
          Import your wallet
        </h1>

        <div className="flex flex-col gap-4 mb-4">
          <div className="flex gap-2 items-center">
            <div className="shrink-0 text-gray-600 dark:text-neutral-400">
              <LifeBuoy className="w-6 h-6" />
            </div>
            <span className="text-gray-600 dark:text-neutral-400">
              Recovery phrase is a set of 12 words that{" "}
              <b>restores your wallet from a backup</b>
            </span>
          </div>
          <div className="flex gap-2 items-center">
            <div className="shrink-0 text-gray-600 dark:text-neutral-400">
              <ShieldCheck className="w-6 h-6" />
            </div>
            <span className="text-gray-600 dark:text-neutral-400">
              Make sure to enter them somewhere safe and private
            </span>
          </div>
        </div>

        <MnemonicInputs mnemonic={mnemonic} setMnemonic={setMnemonic} />
        <Button>Finish</Button>
      </form>
    </>
  );
}
