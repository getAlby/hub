import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { useToast } from "src/components/ui/use-toast";
import useSetupStore from "src/state/SetupStore";

export function LNDForm() {
  const { toast } = useToast();
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [lndAddress, setLndAddress] = React.useState<string>(
    setupStore.nodeInfo.lndAddress || ""
  );
  const [lndCertHex, setLndCertHex] = React.useState<string>(
    setupStore.nodeInfo.lndCertHex || ""
  );
  const [lndMacaroonHex, setLndMacaroonHex] = React.useState<string>(
    setupStore.nodeInfo.lndMacaroonHex || ""
  );

  // TODO: proper onboarding
  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!lndAddress || !lndCertHex || !lndMacaroonHex) {
      toast({
        title: "Please fill out all fields",
        variant: "destructive",
      });
      return;
    }
    handleSubmit({
      lndAddress,
      lndCertHex,
      lndMacaroonHex,
    });
  }

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType: "LND",
      ...data,
    });
    navigate("/setup/finish");
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title="Configure LND"
        description="Fill out wallet details to finish setup."
      />
      <form className="w-full grid gap-5 mt-6" onSubmit={onSubmit}>
        <div className="grid gap-1.5">
          <Label htmlFor="lnd-address">LND Address (GRPC)</Label>
          <Input
            name="lnd-address"
            onChange={(e) => setLndAddress(e.target.value)}
            value={lndAddress}
            id="lnd-address"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="lnd-cert-hex">TLS Certificate (Hex)</Label>
          <Input
            name="lnd-cert-hex"
            onChange={(e) => setLndCertHex(e.target.value)}
            value={lndCertHex}
            type="text"
            id="lnd-cert-hex"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="lnd-macaroon-hex">Admin Macaroon (Hex)</Label>
          <Input
            name="lnd-macaroon-hex"
            onChange={(e) => setLndMacaroonHex(e.target.value)}
            value={lndMacaroonHex}
            type="text"
            id="lnd-macaroon-hex"
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
