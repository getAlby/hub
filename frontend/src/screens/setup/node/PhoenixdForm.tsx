import React from "react";
import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import Container from "src/components/Container";
import PasswordInput from "src/components/password/PasswordInput";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function PhoenixdForm() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [phoenixdAddress, setPhoenixdAddress] = React.useState<string>(
    setupStore.nodeInfo.phoenixdAddress || "http://127.0.0.1:9740"
  );
  const [phoenixdAuthorization, setPhoenixdAuthorization] =
    React.useState<string>(setupStore.nodeInfo.phoenixdAuthorization || "");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!phoenixdAddress || !phoenixdAuthorization) {
      toast.error("Please fill out all fields");
      return;
    }
    handleSubmit({
      phoenixdAddress,
      phoenixdAuthorization,
    });
  }

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType: "PHOENIX",
      ...data,
    });
    navigate("/setup/security");
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title="Configure phoenixd"
        description="Fill out wallet details to finish setup."
      />
      <form className="w-full grid gap-5 mt-6" onSubmit={onSubmit}>
        <div className="grid gap-1.5">
          <Label htmlFor="phoenix-address">Phoenixd Address</Label>
          <Input
            name="phoenix-address"
            onChange={(e) => setPhoenixdAddress(e.target.value)}
            placeholder="http://127.0.0.1:9740"
            value={phoenixdAddress}
            id="phoenix-address"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="lnd-cert-hex">Authorization</Label>

          <PasswordInput
            id="phoenix-authorization"
            onChange={setPhoenixdAuthorization}
            value={phoenixdAuthorization}
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
