import React from "react";
import { ButtonProps, LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";
import { AlbyBalance, Channel } from "src/types";
import { request } from "src/utils/request";

type TransferFundsButtonProps = {
  channels: Channel[] | undefined;
  albyBalance: AlbyBalance;
  onTransferComplete: () => Promise<void>;
} & ButtonProps;

export function TransferFundsButton({
  channels,
  albyBalance,
  onTransferComplete,
  children,
  ...props
}: TransferFundsButtonProps) {
  const [loading, setLoading] = React.useState(false);

  const { toast } = useToast();

  return (
    <LoadingButton
      loading={loading}
      onClick={async () => {
        if (!albyBalance) {
          return;
        }
        if (
          !channels?.some(
            (channel) => channel.remoteBalance / 1000 > albyBalance.sats
          )
        ) {
          toast({
            title: "Please increase your receiving capacity first",
          });
          return;
        }
        setLoading(true);
        try {
          await request("/api/alby/drain", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
          });
          await onTransferComplete();
          toast({
            title:
              "🎉 Funds from Alby shared wallet transferred to your Alby Hub!",
          });
        } catch (error) {
          console.error(error);
          toast({
            variant: "destructive",
            description: "Something went wrong: " + error,
          });
        } finally {
          setLoading(false);
        }
      }}
      {...props}
    >
      {children}
    </LoadingButton>
  );
}
