import { useState } from "react";
import { useNavigate } from "react-router-dom";
import MnemonicInputs from "src/components/mnemonic/MnemonicInputs";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

type Props = {
  open?: boolean;
  defaultOpen?: boolean;
  onOpenChange?(open: boolean): void;
  mnemonic: string;
};

export default function MnemonicDialog({
  open,
  defaultOpen,
  onOpenChange,
  mnemonic,
}: Props) {
  const { data: info } = useInfo();
  const { mutate: refetchInfo } = useInfo();
  const [backedUp, setIsBackedUp] = useState<boolean>(false);
  const [backedUp2, setIsBackedUp2] = useState<boolean>(false);
  const { toast } = useToast();
  const navigate = useNavigate();

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();

    const sixMonthsLater = new Date();
    sixMonthsLater.setMonth(sixMonthsLater.getMonth() + 6);

    try {
      await request("/api/backup-reminder", {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          nextBackupReminder: sixMonthsLater.toISOString(),
        }),
      });
      await refetchInfo();
      navigate("/");
      toast({ title: "Recovery phrase backed up!" });
    } catch (error) {
      handleRequestError(toast, "Failed to store back up info", error);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange} defaultOpen={defaultOpen}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Wallet Recovery Phrase</DialogTitle>
          <DialogDescription>
            Write these words down, store them somewhere safe, and keep them
            secret.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={onSubmit}
          className="flex flex-col gap-2 max-w-md text-sm"
        >
          <MnemonicInputs mnemonic={mnemonic} readOnly={true} asCard={false} />
          <div className="flex flex-col gap-2 mt-4">
            <div className="flex">
              <Checkbox
                id="backup"
                required
                checked={backedUp}
                onCheckedChange={() => setIsBackedUp(!backedUp)}
                className="mt-0.5"
              />
              <Label htmlFor="backup" className="ml-2 text-sm text-foreground">
                I've backed up my recovery phrase to my wallet in a private and
                secure place
              </Label>
            </div>

            {backedUp && !info?.albyAccountConnected && (
              <div className="flex">
                <Checkbox
                  id="backup2"
                  required
                  checked={backedUp2}
                  onCheckedChange={() => setIsBackedUp2(!backedUp2)}
                  className="mt-0.5"
                />
                <Label
                  htmlFor="backup2"
                  className="ml-2 text-sm text-foreground"
                >
                  I understand the recovery phrase AND a backup of my hub data
                  directory after each channel opening is required to recover
                  funds from my lightning channels.
                </Label>
              </div>
            )}
          </div>
          <div className="flex justify-end items-center mt-2">
            <Button type="submit" size="lg" className="grow sm:grow-0">
              Finish
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
