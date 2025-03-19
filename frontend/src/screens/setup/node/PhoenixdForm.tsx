import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import PasswordInput from "src/components/password/PasswordInput";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import useSetupStore from "src/state/SetupStore";

export function PhoenixdForm() {
  const { toast } = useToast();
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
      toast({
        title: "Please fill out all fields",
        variant: "destructive",
      });
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
            name="phoenix-authorization"
            onChange={setPhoenixdAuthorization}
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
