import React from "react";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request";

export function SetupCode() {
  const [setupCode, setSetupCode] = React.useState("");
  const [isLoading, setLoading] = React.useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmedSetupCode = setupCode.trim();
    setLoading(true);
    try {
      await request("/api/setup/verify-code", {
        method: "POST",
        headers: {
          "X-Setup-Code": trimmedSetupCode,
        },
      });
      useSetupStore.getState().setSetupCode(trimmedSetupCode);
    } catch (error) {
      handleRequestError("Failed to verify setup code", error);
      setLoading(false);
    }
  }

  return (
    <Container>
      <form onSubmit={onSubmit} className="flex flex-col items-center w-full">
        <div className="grid gap-4 w-full">
          <TwoColumnLayoutHeader
            title="Enter Setup Code"
            pageTitle="Enter Setup Code"
            description="Enter the setup code that was shown when Alby Hub was installed."
          />
          <div className="grid gap-1.5">
            <Label htmlFor="setup-code">Setup Code</Label>
            <Input
              id="setup-code"
              value={setupCode}
              onChange={(e) => setSetupCode(e.target.value)}
              autoComplete="off"
              autoFocus
              required
            />
          </div>
          <LoadingButton type="submit" loading={isLoading}>
            Continue
          </LoadingButton>
        </div>
      </form>
    </Container>
  );
}
