import React from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApp } from "src/hooks/useApp";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

type IsolatedAppTopupProps = {
  appId: number;
};

export function IsolatedAppTopupDialog({
  appId,
  children,
}: React.PropsWithChildren<IsolatedAppTopupProps>) {
  const { mutate: reloadApp } = useApp(appId);
  const [amountSat, setAmountSat] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [open, setOpen] = React.useState(false);
  const { toast } = useToast();
  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await request(`/api/transfers`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          toAppId: appId,
          amountSat: +amountSat,
        }),
      });
      await reloadApp();
      toast({
        title: `Successfully increased balance by ${+amountSat} sats`,
      });
      reset();
    } catch (error) {
      handleRequestError(toast, "Failed to top up sub-wallet balance", error);
    }
    setLoading(false);
  }

  function reset() {
    setOpen(false);
    setAmountSat("");
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <form onSubmit={onSubmit}>
          <DialogHeader>
            <DialogTitle>Top Up</DialogTitle>
            <DialogDescription>
              Increase the balance of this sub-wallet. Make sure you always
              maintain enough funds in your spending balance to prevent
              sub-wallets becoming unspendable.
            </DialogDescription>
          </DialogHeader>
          <div className="my-5">
            <Label htmlFor="amount">Amount (sats)</Label>
            <Input
              autoFocus
              id="amount"
              type="number"
              required
              value={amountSat}
              onChange={(e) => {
                setAmountSat(e.target.value.trim());
              }}
            />
          </div>
          <DialogFooter className="mt-5">
            <LoadingButton loading={loading}>Top Up</LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
