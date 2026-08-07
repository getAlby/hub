import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english.js";
import {
  AlertTriangleIcon,
  LifeBuoyIcon,
  ShieldAlertIcon,
  ShieldCheckIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router";

import { toast } from "sonner";
import MnemonicInputs from "src/components/mnemonic/MnemonicInputs";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Alert, AlertDescription, AlertTitle } from "src/components/ui/alert";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function ImportMnemonic() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [searchParams] = useSearchParams();
  const nodeParam = (searchParams.get("node") || "").toLowerCase();
  const isGreenlight =
    nodeParam === "greenlight" ||
    setupStore.nodeInfo.backendType === "GREENLIGHT";
  const [backedUp, setIsBackedUp] = useState<boolean>(false);

  useEffect(() => {
    // in case the user presses back, remove their last-saved mnemonic
    useSetupStore.getState().updateNodeInfo({
      mnemonic: undefined,
    });
  }, []);
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

    const currentDate = new Date();
    const sixMonthsLater = new Date(
      currentDate.setMonth(currentDate.getMonth() + 6)
    );

    setupStore.updateNodeInfo({
      mnemonic,
      nextBackupReminder: sixMonthsLater.toISOString(),
    });
    setupStore.setHasImportedMnemonic(true);

    if (isGreenlight) {
      setupStore.updateNodeInfo({ backendType: "GREENLIGHT" });
      navigate("/setup/node/greenlight");
      return;
    }

    navigate(`/setup/node`);
  }

  return (
    <form
      onSubmit={onSubmit}
      className="flex flex-col gap-5 mx-auto max-w-md text-sm"
    >
      <TwoColumnLayoutHeader
        title="Import Recovery Phrase"
        pageTitle="Import Recovery Phrase"
        description={
          isGreenlight
            ? "Enter your 12-word phrase to restore access to your Greenlight node."
            : "Enter your recovery phrase to import your Alby Hub."
        }
      />

      {isGreenlight ? (
        <Alert>
          <LifeBuoyIcon />
          <AlertTitle>Greenlight recovery</AlertTitle>
          <AlertDescription className="inline">
            Your phrase restores Lightning access while Greenlight is online.
            App connections (NWC) are not in the phrase — restore a Hub backup
            if you have one, or create connections again after unlock.
          </AlertDescription>
        </Alert>
      ) : (
        <Alert variant="warning">
          <AlertTriangleIcon />
          <AlertTitle>Do not re-use the same key on multiple devices</AlertTitle>
          <AlertDescription className="inline">
            If you want to transfer your existing Hub to another machine please
            use the <b>Hub backup</b> option from restore.
          </AlertDescription>
        </Alert>
      )}
      <Alert className="grid-cols-none">
        <div className="flex flex-col gap-4">
          <div className="flex gap-2 items-center">
            <div className="shrink-0 text-muted-foreground">
              <LifeBuoyIcon className="size-6" />
            </div>
            <span className="text-muted-foreground">
              {isGreenlight
                ? "Your recovery phrase is 12 words. It is the same phrase Greenlight uses for your node seed."
                : "Your recovery phrase is a set of 12 words used to restore your on-chain balance from a backup."}
            </span>
          </div>
          <div className="flex gap-2 items-center">
            <div className="shrink-0 text-muted-foreground">
              <ShieldCheckIcon className="size-6" />
            </div>
            <span className="text-muted-foreground">
              Keep it safe and private to ensure your funds remain secure.
            </span>
          </div>
          <div className="flex gap-2 items-center">
            <div className="shrink-0 text-muted-foreground">
              <ShieldAlertIcon className="size-6" />
            </div>
            <span className="text-muted-foreground">
              {isGreenlight
                ? "Channels stay on Greenlight. The phrase recovers access to that node, not a Hub backup file."
                : "Your recovery phrase cannot restore funds from lightning channels. If you had active channels on a different device, contact Alby support before proceeding."}
            </span>
          </div>
        </div>
      </Alert>

      <MnemonicInputs mnemonic={mnemonic} setMnemonic={setMnemonic} />

      {!isGreenlight && (
        <div className="flex items-center mt-5">
          <Checkbox
            id="confirmedNoChannels"
            required
            onCheckedChange={() => setIsBackedUp(!backedUp)}
          />
          <Label htmlFor="confirmedNoChannels" className="ml-2 cursor-pointer">
            I don't have another Alby Hub to migrate or open channels (funds from
            channels will be lost!).
          </Label>
        </div>
      )}
      {isGreenlight && (
        <div className="flex items-center mt-5">
          <Checkbox
            id="confirmedGlRecover"
            required
            onCheckedChange={() => setIsBackedUp(!backedUp)}
          />
          <Label htmlFor="confirmedGlRecover" className="ml-2 cursor-pointer">
            I understand app connections will need to be set up again unless I
            restore a Hub backup instead.
          </Label>
        </div>
      )}
      <Button>Next</Button>
    </form>
  );
}
