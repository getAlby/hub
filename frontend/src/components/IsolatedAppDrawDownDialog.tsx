import React from "react";
import { toast } from "sonner";
import { LoadingButton } from "src/components/ui/custom/loading-button";
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
import { useApp } from "src/hooks/useApp";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

type IsolatedAppTopupProps = {
  appId: number;
};

export function IsolatedAppDrawDownDialog({
  appId,
  children,
}: React.PropsWithChildren<IsolatedAppTopupProps>) {
  const { mutate: reloadApp } = useApp(appId);
  const [amountSat, setAmountSat] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [open, setOpen] = React.useState(false);
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
          fromAppId: appId,
          amountSat: +amountSat,
          description,
        }),
      });
      await reloadApp();
      toast(`Successfully reduced balance by ${+amountSat} sats`);
      reset();
    } catch (error) {
      handleRequestError("Failed to decrease sub-wallet balance", error);
    }
    setLoading(false);
  }

  function reset() {
    setOpen(false);
    setAmountSat("");
    setDescription("");
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        <form onSubmit={onSubmit}>
          <DialogHeader>
            <DialogTitle>Decrease Balance</DialogTitle>
            <DialogDescription>
              Decrease the balance of this sub-wallet.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-2 mt-5">
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
          <div className="grid gap-2 mt-3">
            <Label htmlFor="description">Description (optional)</Label>
            <Input
              id="description"
              type="text"
              placeholder="transfer"
              value={description}
              onChange={(e) => {
                setDescription(e.target.value);
              }}
            />
          </div>
          <DialogFooter className="mt-5">
            <LoadingButton loading={loading}>Decrease</LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
