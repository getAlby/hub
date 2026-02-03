import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function CLNForm() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [clnAddress, setClnAddress] = React.useState<string>(
    setupStore.nodeInfo.clnAddress || ""
  );
  const [clnLightningDir, setClnLightningDir] = React.useState<string>(
    setupStore.nodeInfo.clnLightningDir || ""
  );
  const [clnAddressHold, setClnAddressHold] = React.useState<string>(
    setupStore.nodeInfo.clnAddressHold || ""
  );

  // TODO: proper onboarding
  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    handleSubmit({
      clnAddress,
      clnLightningDir,
      clnAddressHold,
    });
  }

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType: "CLN",
      ...data,
    });
    navigate("/setup/security");
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title="Configure CLN"
        description="Fill out wallet details to finish setup."
      />
      <form className="w-full grid gap-5 mt-6" onSubmit={onSubmit}>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-address">CLN Address (GRPC)</Label>
          <Input
            required
            name="cln-address"
            onChange={(e) => setClnAddress(e.target.value)}
            value={clnAddress}
            id="cln-address"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-lightning-dir">
            CLN Lightning directory (full path)
          </Label>
          <Input
            required
            name="cln-lightning-dir"
            onChange={(e) => setClnLightningDir(e.target.value)}
            value={clnLightningDir}
            type="text"
            id="cln-lightning-dir"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-address-hold">
            (optional) CLN hold plugin Address (GRPC)
          </Label>
          <Input
            name="cln-address-hold"
            onChange={(e) => setClnAddressHold(e.target.value)}
            value={clnAddressHold}
            id="cln-address-hold"
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
