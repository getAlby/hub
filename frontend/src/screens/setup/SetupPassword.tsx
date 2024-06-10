import React, { useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import { useInfo } from "src/hooks/useInfo";

export function SetupPassword() {
  const navigate = useNavigate();
  const store = useSetupStore();
  const { toast } = useToast();
  const { data: info } = useInfo();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [isPasswordSecured, setIsPasswordSecured] = useState<boolean>(false);

  const [searchParams] = useSearchParams();
  const wallet = searchParams.get("wallet") || "new";
  const node = searchParams.get("node") || "";

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

    if (wallet === "import") {
      navigate(`/setup/import-mnemonic`);
    } else if (node) {
      navigate(`/setup/node/${node}`);
    } else {
      navigate(`/setup/node`);
    }
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
              <div className="grid gap-1.5">
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
              <div className="grid gap-1.5">
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
          </div>
        </form>
      </div>
    </>
  );
}
