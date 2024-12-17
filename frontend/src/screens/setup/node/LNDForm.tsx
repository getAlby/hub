import { InfoIcon } from "lucide-react";
import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function LNDForm() {
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
            required
            name="lnd-address"
            onChange={(e) => setLndAddress(e.target.value)}
            value={lndAddress}
            id="lnd-address"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="lnd-macaroon-hex">Admin Macaroon (Hex)</Label>
          <Input
            required
            name="lnd-macaroon-hex"
            onChange={(e) => setLndMacaroonHex(e.target.value)}
            value={lndMacaroonHex}
            type="text"
            id="lnd-macaroon-hex"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="lnd-cert-hex">TLS Certificate (Hex) (optional)</Label>
          <Input
            name="lnd-cert-hex"
            onChange={(e) => setLndCertHex(e.target.value)}
            value={lndCertHex}
            type="text"
            id="lnd-cert-hex"
          />
          {!lndCertHex && (
            <div className="flex flex-row gap-2 items-center justify-start text-sm text-muted-foreground mt-2">
              <InfoIcon className="h-4 w-4 shrink-0" />
              Skipping TLS certificate is not recommended as it may expose your
              connection to security risks
            </div>
          )}
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
