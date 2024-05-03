import React from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";

export function SetupPassword() {
  const { toast } = useToast();
  const store = useSetupStore();
  const { data: info } = useInfo();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (store.unlockPassword !== confirmPassword) {
      toast({
        title: "Passwords don't match",
        variant: "destructive",
      });
      return;
    }

    const wallet = searchParams.get("wallet");

    // Pre-configured LND
    if (info?.backendType === "LND") {
      // NOTE: LND flow does not setup a mnemonic
      navigate(`/setup/finish`);
      return;
    }

    // Import flow (All options)
    if (wallet === "import") {
      navigate(`/setup/node`);
      return;
    }

    // Default flow (LDK)
    useSetupStore.getState().updateNodeInfo({
      backendType: "LDK",
      mnemonic: bip39.generateMnemonic(wordlist, 128),
    });
    navigate(`/setup/finish`);
  }

  return (
    <>
      <Container>
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <div className="grid gap-5">
            <TwoColumnLayoutHeader
              title="Create Password"
              description="You'll use it to access your Alby Hub on any device."
            />
            <div className="grid gap-4 w-full">
              <div className="grid gap-1.5">
                <Label htmlFor="unlock-password">New Password</Label>
                <Input
                  type="password"
                  name="unlock-password"
                  id="unlock-password"
                  autoComplete="new-password"
                  placeholder="Enter a password"
                  value={store.unlockPassword}
                  onChange={(e) => store.setUnlockPassword(e.target.value)}
                  required={true}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="confirm-password">Confirm Password</Label>
                <Input
                  type="password"
                  name="confirm-password"
                  id="confirm-password"
                  autoComplete="new-password"
                  placeholder="Re-enter the password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  required={true}
                />
              </div>
            </div>
            <Button type="submit">Create Password</Button>
          </div>
        </form>
      </Container>
    </>
  );
}
