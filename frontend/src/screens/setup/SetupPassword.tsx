import React, { useState } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";
import { backendTypeHasMnemonic } from "src/lib/utils";

export function SetupPassword() {
  const { toast } = useToast();
  const store = useSetupStore();
  const { data: info } = useInfo();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const wallet = searchParams.get("wallet");
  const [isPasswordSecured, setIsPasswordSecured] = useState<boolean>(false);

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!info) {
      return;
    }
    if (store.unlockPassword !== confirmPassword) {
      toast({
        title: "Passwords don't match",
        variant: "destructive",
      });
      return;
    }

    if (!backendTypeHasMnemonic(info.backendType)) {
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
      <div className="grid max-w-sm">
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <div className="grid gap-4">
            <TwoColumnLayoutHeader
              title="Create Password"
              description="Your password is used to access your wallet, and it can't be reset or recovered if you lose it."
            />
            <div className="grid gap-4 w-full">
              <div className="grid gap-2">
                <Label htmlFor="unlock-password">Password</Label>
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
              <div className="grid gap-2">
                <Label htmlFor="confirm-password">Repeat Password</Label>
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
            <div className="grid gap-6">
              <div className="flex items-center">
                <Checkbox
                  id="securePassword"
                  onCheckedChange={() =>
                    setIsPasswordSecured(!isPasswordSecured)
                  }
                />
                <Label
                  htmlFor="securePassword"
                  className="ml-2 text-foreground leading-4"
                >
                  I've secured this password in a safe place
                </Label>
              </div>
              <Button type="submit" disabled={!isPasswordSecured}>
                Create Password
              </Button>
            </div>

            {wallet === "import" && (
              <div className="flex flex-col justify-center items-center gap-4">
                <p className="text-muted-foreground">or</p>
                <Link to="/setup/node-restore" className="w-full">
                  <Button variant="secondary" className="w-full">
                    Import Backup File
                  </Button>
                </Link>
              </div>
            )}
          </div>
        </form>
      </div>
    </>
  );
}
