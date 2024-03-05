import { PopiconsLifebuoyLine, PopiconsShieldLine } from "@popicons/react";
import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import MnemonicInputs from "src/components/MnemonicInputs";
import ConnectButton from "src/components/ConnectButton";
import useSetupStore from "src/state/SetupStore";
import toast from "src/components/Toast";

export function SetupMnemonic() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();

  const [mnemonic, setMnemonic] = useState<string>("");

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (
      mnemonic.split(" ").length !== 12 ||
      !bip39.validateMnemonic(mnemonic, wordlist)
    ) {
      toast.error("Invalid recovery phrase");
      return;
    }

    function dateToString(date: Date) {
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, "0");
      const day = String(date.getDate()).padStart(2, "0");
      const hours = String(date.getHours()).padStart(2, "0");
      const minutes = String(date.getMinutes()).padStart(2, "0");
      const seconds = String(date.getSeconds()).padStart(2, "0");

      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    }

    setupStore.updateNodeInfo({
      mnemonic,
      isMnemonicBackupDone: dateToString(new Date()),
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
              <PopiconsLifebuoyLine className="w-6 h-6" />
            </div>
            <span className="text-gray-600 dark:text-neutral-400">
              Recovery phrase is a set of 12 words that{" "}
              <b>restores your wallet from a backup</b>
            </span>
          </div>
          <div className="flex gap-2 items-center">
            <div className="shrink-0 text-gray-600 dark:text-neutral-400">
              <PopiconsShieldLine className="w-6 h-6" />
            </div>
            <span className="text-gray-600 dark:text-neutral-400">
              Make sure to enter them somewhere safe and private
            </span>
          </div>
        </div>

        <MnemonicInputs mnemonic={mnemonic} setMnemonic={setMnemonic} />
        <ConnectButton
          submitText="Finish"
          loadingText="Saving..."
          isConnecting={false}
          disabled={false}
        />
      </form>
    </>
  );
}
