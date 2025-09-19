import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import {
  AlertTriangleIcon,
  LifeBuoyIcon,
  ShieldAlertIcon,
  ShieldCheckIcon,
} from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

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

    navigate(`/setup/security`);
  }

  return (
    <>
      <form
        onSubmit={onSubmit}
        className="flex flex-col gap-5 mx-auto max-w-md text-sm"
      >
        <TwoColumnLayoutHeader
          title="Import Master Key"
          description="Enter the your Master Key recovery phrase to import your Alby Hub."
        />

        <Alert variant="warning">
          <AlertTriangleIcon />
          <AlertTitle>
            Do not re-use the same key on multiple devices
          </AlertTitle>
          <AlertDescription className="inline">
            If you want to transfer your existing Hub to another machine please
            use the <b>migrate feature</b> from the Alby Hub settings.
          </AlertDescription>
        </Alert>
        <Alert className="grid-cols-none">
          <div className="flex flex-col gap-4">
            <div className="flex gap-2 items-center">
              <div className="shrink-0 text-muted-foreground">
                <LifeBuoyIcon className="size-6" />
              </div>
              <span className="text-muted-foreground">
                Your recovery phrase is a set of 12 words used to restore your
                on-chain balance from a backup.
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
                Your recovery phrase cannot restore funds from lightning
                channels. If you had active channels on a different device,
                contact Alby support before proceeding.
              </span>
            </div>
          </div>
        </Alert>

        <MnemonicInputs mnemonic={mnemonic} setMnemonic={setMnemonic} />

        <div className="flex items-center mt-5">
          <Checkbox
            id="confirmedNoChannels"
            required
            onCheckedChange={() => setIsBackedUp(!backedUp)}
          />
          <Label htmlFor="confirmedNoChannels" className="ml-2">
            I don't have another Alby Hub to migrate or open channels (funds
            from channels will be lost!).
          </Label>
        </div>
        <Button>Next</Button>
      </form>
    </>
  );
}
