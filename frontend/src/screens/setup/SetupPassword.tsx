import React, { useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import { toast } from "sonner";
import PasswordInput from "src/components/password/PasswordInput";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Checkbox } from "src/components/ui/checkbox";
import { Label } from "src/components/ui/label";
import { useInfo } from "src/hooks/useInfo";

export function SetupPassword() {
  const navigate = useNavigate();
  const store = useSetupStore();
  const { data: info } = useInfo();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const [isPasswordSecured, setIsPasswordSecured] = useState<boolean>(false);
  const [isPasswordSecured2, setIsPasswordSecured2] = useState<boolean>(false);

  const [searchParams] = useSearchParams();
  const wallet = searchParams.get("wallet") || "new";
  const node = searchParams.get("node") || "";

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!info) {
      return;
    }
    if (!isPasswordSecured || !isPasswordSecured2) {
      toast.error("Please confirm you have saved your password");
      return;
    }
    if (store.unlockPassword !== confirmPassword) {
      toast.error("Passwords don't match");
      return;
    }

    if (wallet === "import") {
      useSetupStore.getState().updateNodeInfo({
        backendType: "LDK",
      });
      navigate(`/setup/import-mnemonic`);
    } else if (node) {
      navigate(`/setup/node/${node}`);
    } else {
      navigate(`/setup/node`);
    }
  }

  return (
    <>
      <title>Create Password - Alby Hub</title>
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
                <PasswordInput
                  id="unlock-password"
                  onChange={store.setUnlockPassword}
                  autoComplete="new-password"
                  placeholder="Enter a password"
                  autoFocus
                  value={store.unlockPassword}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="confirm-password">Repeat Password</Label>
                <PasswordInput
                  id="confirm-password"
                  autoComplete="new-password"
                  placeholder="Re-enter the password"
                  onChange={setConfirmPassword}
                  value={confirmPassword}
                />
              </div>
            </div>
            <div className="grid gap-6">
              <div className="flex items-center">
                <Checkbox
                  id="securePassword"
                  required
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
              {isPasswordSecured && (
                <div className="flex items-center">
                  <Checkbox
                    id="securePassword2"
                    required
                    onCheckedChange={() =>
                      setIsPasswordSecured2(!isPasswordSecured2)
                    }
                  />
                  <Label
                    htmlFor="securePassword2"
                    className="ml-2 leading-4 font-semibold"
                  >
                    I understand this password cannot be recovered
                  </Label>
                </div>
              )}
              <Button type="submit">Create Password</Button>
            </div>
          </div>
        </form>
      </div>
    </>
  );
}
