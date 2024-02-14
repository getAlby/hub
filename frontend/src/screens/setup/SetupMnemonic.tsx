import {
  PopiconsLifebuoyLine,
  PopiconsShieldLine,
  PopiconsTriangleExclamationLine,
} from "@popicons/react";
import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

import MnemonicInputs from "src/components/MnemonicInputs";
import ConnectButton from "src/components/ConnectButton";
import useSetupStore from "src/state/SetupStore";
import toast from "src/components/Toast";

export function SetupMnemonic() {
  const navigate = useNavigate();
  const { search } = useLocation();
  const { updateNodeInfo } = useSetupStore();
  const params = new URLSearchParams(search);
  const isNew = params.get("wallet") === "new";

  const [mnemonic, setMnemonic] = useState<string>(
    isNew ? bip39.generateMnemonic(wordlist, 128) : ""
  );
  const [backedUp, isBackedUp] = useState<boolean>(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (
      mnemonic.split(" ").length !== 12 ||
      !bip39.validateMnemonic(mnemonic, wordlist)
    ) {
      toast.error("Invalid recovery phrase");
      return;
    }

    updateNodeInfo({ mnemonic });
    navigate(`/setup/finish`);
  }

  return (
    <>
      <form
        onSubmit={onSubmit}
        className="flex flex-col gap-2 mx-auto max-w-2xl text-sm"
      >
        <h1 className="font-semibold text-2xl font-headline mb-2 dark:text-white">
          {isNew ? "Back up your wallet" : "Import your wallet"}
        </h1>

        {isNew ? (
          <div className="flex flex-col gap-4 mb-4">
            <div className="flex gap-2 items-center">
              <div className="shrink-0 text-gray-600 dark:text-neutral-400">
                <PopiconsLifebuoyLine className="w-6 h-6" />
              </div>
              <span className="text-gray-600 dark:text-neutral-400">
                Your recovery phrase is a set of 12 words that{" "}
                <b>backs up your wallet</b>
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <div className="shrink-0 text-gray-600 dark:text-neutral-400">
                <PopiconsShieldLine className="w-6 h-6" />
              </div>
              <span className="text-gray-600 dark:text-neutral-400">
                Make sure to write them down somewhere safe and private
              </span>
            </div>
            <div className="flex gap-2 items-center">
              <div className="shrink-0 text-red-600 dark:text-red-800">
                <PopiconsTriangleExclamationLine className="w-6 h-6" />
              </div>
              <span className="font-medium text-red-600 dark:text-red-800">
                If you lose your recovery phrase, you will lose access to your
                funds
              </span>
            </div>
          </div>
        ) : (
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
        )}

        <MnemonicInputs
          mnemonic={mnemonic}
          setMnemonic={setMnemonic}
          readOnly={isNew}
        >
          {isNew && (
            <div className="flex items-center">
              <input
                id="checkbox"
                type="checkbox"
                onChange={(event) => {
                  isBackedUp(event.target.checked);
                }}
                checked={backedUp}
                className="w-4 h-4 text-purple-700 bg-gray-100 border-gray-300 rounded focus:ring-purple-700 dark:focus:ring-purple-800 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
              />
              <label
                htmlFor="checkbox"
                className="ms-2 text-sm font-medium text-gray-900 dark:text-gray-100"
              >
                I've backed my recovery phrase to my wallet in a private and
                secure place
              </label>
            </div>
          )}
        </MnemonicInputs>
        <ConnectButton
          submitText="Finish"
          loadingText="Saving..."
          isConnecting={false}
          disabled={isNew ? !backedUp : false}
        />
      </form>
    </>
  );
}
