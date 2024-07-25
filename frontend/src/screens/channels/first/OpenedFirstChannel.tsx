import confetti from "canvas-confetti";
import React from "react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { useToast } from "src/components/ui/use-toast";
import { ALBY_HIDE_HOSTED_BALANCE_BELOW } from "src/constants";
import { useAlbyBalance } from "src/hooks/useAlbyBalance";
import { useCSRF } from "src/hooks/useCSRF";
import { request } from "src/utils/request";

export function OpenedFirstChannel() {
  const { data: csrf } = useCSRF();
  const { data: albyBalance, mutate: reloadAlbyBalance } = useAlbyBalance();
  const [, setShowedAlbyMigrationToast] = React.useState(false);
  const { toast } = useToast();

  // automatically drain Alby balance into new channel if possible
  // TODO: remove this code once all Alby users have migrated to Alby Hub
  React.useEffect(() => {
    (async () => {
      if (
        !csrf ||
        !albyBalance ||
        albyBalance.sats < ALBY_HIDE_HOSTED_BALANCE_BELOW
      ) {
        return;
      }
      try {
        await request("/api/alby/drain", {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
        });
        await reloadAlbyBalance();
        // This may run multiple times (to drain the final 1%), but we should only show a toast once
        setShowedAlbyMigrationToast((current) => {
          if (!current) {
            toast({
              description:
                "🎉 Funds from Alby shared wallet transferred to your Alby Hub!",
            });
          }
          return true;
        });
      } catch (error) {
        console.error("Failed to transfer any alby shared wallet funds", error);
      }
    })();
  }, [albyBalance, csrf, reloadAlbyBalance, toast]);

  React.useEffect(() => {
    for (let i = 0; i < 10; i++) {
      setTimeout(
        () => {
          confetti({
            origin: {
              x: Math.random(),
              y: Math.random(),
            },
            colors: ["#000", "#333", "#666", "#999", "#BBB", "#FFF"],
          });
        },
        Math.floor(Math.random() * 1000)
      );
    }
  }, []);

  return (
    <div className="flex flex-col justify-center gap-5 p-5 max-w-md items-stretch">
      <TwoColumnLayoutHeader
        title="Channel Opened"
        description="Your new lightning channel is ready to use"
      />

      <p>
        Congratulations! Your first lightning channel is active and can be used
        to send and receive payments.
      </p>
      <p>
        To ensure you can both send and receive, make sure to balance your{" "}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/liquidity"
          className="underline"
        >
          channel's liquidity
        </ExternalLink>
        .
      </p>

      <Link to="/home" className="flex justify-center mt-8">
        <Button>Go to your dashboard</Button>
      </Link>
    </div>
  );
}
