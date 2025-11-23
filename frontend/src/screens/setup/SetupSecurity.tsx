import {
  HandCoinsIcon,
  LandmarkIcon,
  ShieldAlertIcon,
  UnlockIcon,
} from "lucide-react";
import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";

import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";

export function SetupSecurity() {
  const navigate = useNavigate();
  const [hasConfirmed, setConfirmed] = useState<boolean>(false);
  const store = useSetupStore();
  const { data: info } = useInfo();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    navigate("/setup/finish");
  }

  return (
    <>
      <div className="grid max-w-sm">
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <TwoColumnLayoutHeader
            title="Security & Recovery"
            description="Take your time to understand how to secure and recover your funds on Alby Hub."
          />

          <div className="flex flex-col gap-6 w-full mt-6">
            {store.nodeInfo.backendType !== "CASHU" && (
              <div className="flex gap-3 items-center">
                <div className="shrink-0">
                  <HandCoinsIcon className="size-6" />
                </div>
                <span className="text-sm text-muted-foreground">
                  Alby Hub is a spending wallet - do not keep all your savings
                  on it!
                </span>
              </div>
            )}
            {store.nodeInfo.backendType === "CASHU" && (
              <div className="flex gap-3 items-center">
                <div className="shrink-0">
                  <LandmarkIcon className="size-6" />
                </div>
                <span className="text-sm text-muted-foreground">
                  Your funds are owned by the cashu mint - use at your own risk
                  with only small amounts!
                </span>
              </div>
            )}
            <div className="flex gap-3 items-center">
              <div className="shrink-0">
                <UnlockIcon className="size-6" />
              </div>
              <span className="text-sm text-muted-foreground">
                Access to your Alby Hub is protected by an unlock password you
                set. It cannot be recovered or reset.
              </span>
            </div>
            {store.nodeInfo.backendType === "LND" ||
            store.nodeInfo.backendType === "PHOENIX" ? (
              <div className="flex gap-3 items-center">
                <div className="shrink-0">
                  <ShieldAlertIcon className="size-6" />
                </div>
                <span className="text-sm text-muted-foreground">
                  Channel backups{" "}
                  <span className="underline">are not handled</span> by Alby
                  Hub. Please take care of your own backups or go back and
                  choose the LDK node type.
                </span>
              </div>
            ) : (
              <div className="flex gap-3 items-center">
                <div className="shrink-0">
                  <ShieldAlertIcon className="size-6" />
                </div>
                <span className="text-sm text-muted-foreground">
                  Your{store.nodeInfo.backendType === "LDK" && " on-chain"}{" "}
                  balance can be recovered only with your 12-word recovery
                  phrase which you can access once your node has started.
                  {!info?.albyUserIdentifier &&
                    store.nodeInfo.backendType === "LDK" && (
                      <> You must also take care of your own channel backups.</>
                    )}
                </span>
              </div>
            )}
            <ExternalLink
              className="text-muted-foreground flex items-center text-sm"
              to="https://guides.getalby.com/user-guide/alby-hub/backups-and-recover"
            >
              <p>
                Learn more about backups and recovery process on{" "}
                <span className="font-semibold underline">Alby Guides</span>.
              </p>
            </ExternalLink>
            <div className="flex items-center">
              <Checkbox
                id="securePassword"
                required
                onCheckedChange={() => setConfirmed(!hasConfirmed)}
              />
              <Label
                htmlFor="securePassword"
                className="ml-2 text-foreground leading-4"
              >
                I understand how to secure and recover funds
              </Label>
            </div>
            <Button className="w-full" disabled={!hasConfirmed} type="submit">
              Continue
            </Button>
          </div>
        </form>
      </div>
    </>
  );
}
