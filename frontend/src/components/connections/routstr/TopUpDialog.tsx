import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { fundFromHub, getRoutstrdKeyBalances } from "src/hooks/useRoutstrd";
import { handleRequestError } from "src/utils/handleRequestError";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onTopUp: (amount: number) => void;
  /** Routstr connection app ID — sats come from this app's isolated wallet */
  appId: number;
};

export function TopUpDialog({ open, onOpenChange, onTopUp, appId }: Props) {
  const [amount, setAmount] = useState("1000");
  const [isProcessing, setIsProcessing] = useState(false);

  useEffect(() => {
    if (open) {
      setAmount("1000");
      setIsProcessing(false);
    }
  }, [open]);

  const handleTopUp = async () => {
    const numAmount = Number(amount);
    if (!Number.isInteger(numAmount) || numAmount < 1) {
      toast.error("Enter a valid amount");
      return;
    }

    setIsProcessing(true);
    // Snapshot balance to confirm success
    let before = 0;
    try {
      const bal = await getRoutstrdKeyBalances();
      before = bal?.total ?? 0;
    } catch {
      /* ignore */
    }

    try {
      // Direct Hub-to-mint path — pay from this Routstr connection's wallet
      await fundFromHub(numAmount, appId);
      onTopUp(numAmount);
      onOpenChange(false);
      toast.success(`Deposited ${numAmount} sats`);
    } catch (error) {
      // If Hub path failed, check if balance increased anyway
      try {
        await new Promise((r) => setTimeout(r, 2000));
        const afterBal = await getRoutstrdKeyBalances();
        const after = afterBal?.total ?? 0;
        if (after >= before + numAmount) {
          onTopUp(numAmount);
          onOpenChange(false);
          toast.success(`Deposited ${numAmount} sats`);
          return;
        }
      } catch {
        /* fall through */
      }
      handleRequestError("Deposit failed", error);
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Top Up API Balance</DialogTitle>
        </DialogHeader>

        <div className="py-4">
          <Label htmlFor="topup-amount">Amount (sats)</Label>
          <Input
            id="topup-amount"
            type="number"
            min={1}
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="mt-2"
          />
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <LoadingButton
            loading={isProcessing}
            onClick={handleTopUp}
            disabled={!amount || Number(amount) < 1}
          >
            Deposit {amount} sats
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
