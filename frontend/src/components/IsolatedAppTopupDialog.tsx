import React from "react";
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Input } from "src/components/ui/input";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { useApp } from "src/hooks/useApp";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

type IsolatedAppTopupProps = {
  appPubkey: string;
};

export function IsolatedAppTopupDialog({
  appPubkey,
  children,
}: React.PropsWithChildren<IsolatedAppTopupProps>) {
  const { mutate: reloadApp } = useApp(appPubkey);
  const [amountSat, setAmountSat] = React.useState("");
  const [loading, setLoading] = React.useState(false);
  const [open, setOpen] = React.useState(false);
  const { toast } = useToast();
  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    try {
      await request(`/api/apps/${appPubkey}/topup`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          amountSat: +amountSat,
        }),
      });
      await reloadApp();
      toast({
        title: "Successfully increased isolated app balance",
      });
      setOpen(false);
    } catch (error) {
      handleRequestError(
        toast,
        "Failed to increase isolated app balance",
        error
      );
    }
    setLoading(false);
  }
  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>{children}</AlertDialogTrigger>
      <AlertDialogContent>
        <form onSubmit={onSubmit}>
          <AlertDialogHeader>
            <AlertDialogTitle>Increase Isolated App Balance</AlertDialogTitle>
            <AlertDialogDescription>
              As the owner of your Alby Hub, you must make sure you have enough
              funds in your channels for this app to make payments matching its
              balance.
            </AlertDialogDescription>
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
          </AlertDialogHeader>
          <AlertDialogFooter className="mt-5">
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <LoadingButton loading={loading}>Top Up</LoadingButton>
          </AlertDialogFooter>
        </form>
      </AlertDialogContent>
    </AlertDialog>
  );
}
