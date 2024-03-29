import React from "react";
import { useNavigate } from "react-router-dom";
import useSetupStore from "src/state/SetupStore";

import Container from "src/components/Container";
import toast from "src/components/Toast";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";

export function SetupPassword() {
  const store = useSetupStore();
  const [confirmPassword, setConfirmPassword] = React.useState("");
  const navigate = useNavigate();

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (store.unlockPassword !== confirmPassword) {
      toast.error("Passwords don't match!");
      return;
    }
    navigate("/setup/wallet");
  }

  return (
    <>
      <Container>
        <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
          <div className="grid gap-5">
            <div className="grid gap-2 text-center">
              <h1 className="font-semibold text-2xl font-headline">
                Create Password
              </h1>
              <p className="text-muted-foreground">
                You’ll use it to access your Alby Hub on any device.
              </p>
            </div>
            <div className="grid gap-3 w-full">
              <div className="grid gap-1.5">
                <Label htmlFor="unlock-password">New Password</Label>
                <Input
                  type="password"
                  name="unlock-password"
                  id="unlock-password"
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
