import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "src/components/ui/dialog";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import {
  deleteRoutstrdClient,
  getRoutstrdKeyBalances,
} from "src/hooks/useRoutstrd";
import { handleRequestError } from "src/utils/handleRequestError";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clientId: string;
  onDeleted: () => void;
};

export function DeleteKeyDialog({
  open,
  onOpenChange,
  clientId,
  onDeleted,
}: Props) {
  const [balance, setBalance] = useState<number | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [daemonUnreachable, setDaemonUnreachable] = useState(false);

  // Wallet holds sats — not the API key string. Use total Cashu balance.
  const checkBalance = useCallback(async () => {
    try {
      const result = await getRoutstrdKeyBalances();
      const total =
        typeof result?.total === "number"
          ? result.total
          : (result?.keys || []).reduce((sum, k) => sum + (k.balance || 0), 0);
      setBalance(total);
      setDaemonUnreachable(false);
    } catch (error) {
      // Distinguish a confirmed "client not found" (safe to delete — the
      // handleDelete path also treats it as success) from a transient
      // failure (daemon restarting, network blip). Treating a transient
      // failure as zero balance would let the key be deleted while it
      // still holds sats, stranding them.
      const message = error instanceof Error ? error.message : String(error);
      if (/not found/i.test(message)) {
        setBalance(0);
      } else {
        setBalance(null);
      }
      setDaemonUnreachable(true);
    }
  }, []);

  useEffect(() => {
    if (open) {
      setBalance(null);
      setIsProcessing(false);
      setDaemonUnreachable(false);
      checkBalance();
    }
  }, [open, checkBalance]);

  const handleDelete = async () => {
    if (balance !== null && balance > 0) {
      toast.error(
        `Wallet still has ${balance} sats. Refund to zero before deleting the key.`
      );
      return;
    }

    setIsProcessing(true);
    try {
      await deleteRoutstrdClient(clientId);
      toast.success("API key deleted");
      onDeleted();
      onOpenChange(false);
    } catch (error) {
      // Daemon may already have dropped this client (direct deletion, DB
      // restore, wizard abandoned). "not found" = already gone = success —
      // otherwise the ghost key + Delete button stay stuck forever (404 loop).
      const message = error instanceof Error ? error.message : String(error);
      if (/not found/i.test(message)) {
        toast.success("API key already removed");
        onDeleted();
        onOpenChange(false);
      } else {
        handleRequestError("Failed to delete API key", error);
      }
    } finally {
      setIsProcessing(false);
    }
  };

  const canDelete = balance !== null && balance === 0;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>Delete API Key</DialogTitle>
          <DialogDescription>
            Removes this key from the local Routstr daemon. Refund any remaining
            wallet balance first so sats are not stranded.
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-lg border border-border/60 bg-muted/10 p-4 space-y-2 text-sm">
          {balance === null ? (
            <div className="text-muted-foreground">
              Checking wallet balance…
            </div>
          ) : (
            <>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Wallet balance</span>
                <span className="font-medium tabular-nums">{balance} sats</span>
              </div>
              {balance > 0 && (
                <p className="text-xs text-amber-600 dark:text-amber-400">
                  Refund remaining sats to your Routstr NWC wallet before
                  deleting.
                </p>
              )}
              {daemonUnreachable && (
                <p className="text-xs text-muted-foreground">
                  Daemon unreachable — deletion allowed to clear Hub metadata.
                </p>
              )}
              {balance === 0 && !daemonUnreachable && (
                <p className="text-xs text-muted-foreground">
                  Balance is zero. Safe to delete.
                </p>
              )}
            </>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <LoadingButton
            loading={isProcessing}
            variant="destructive"
            onClick={handleDelete}
            disabled={!canDelete}
          >
            Delete Key
          </LoadingButton>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
