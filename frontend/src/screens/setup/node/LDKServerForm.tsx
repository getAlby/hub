import React from "react";
import { useNavigate } from "react-router";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import useSetupStore from "src/state/SetupStore";

export function LDKServerForm() {
  const navigate = useNavigate();
  const setupStore = useSetupStore();
  const [ldkServerAddress, setLdkServerAddress] = React.useState<string>(
    setupStore.nodeInfo.ldkServerAddress || "127.0.0.1:3536"
  );
  const [ldkServerTlsCertFile, setLdkServerTlsCertFile] = React.useState(
    setupStore.nodeInfo.ldkServerTlsCertFile || ""
  );
  const [ldkServerApiKey, setLdkServerApiKey] = React.useState<string>(
    setupStore.nodeInfo.ldkServerApiKey || ""
  );

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setupStore.updateNodeInfo({
      backendType: "LDK_SERVER",
      ldkServerAddress,
      ldkServerTlsCertFile,
      ldkServerApiKey,
    });
    navigate("/setup/security");
  }

  return (
    <Container>
      <TwoColumnLayoutHeader
        title="Configure LDK Server"
        description="Connect Hub to an existing ldk-server gRPC endpoint."
      />
      <form className="w-full grid gap-5 mt-6" onSubmit={onSubmit}>
        <div className="grid gap-1.5">
          <Label htmlFor="ldk-server-address">gRPC Address</Label>
          <Input
            required
            id="ldk-server-address"
            value={ldkServerAddress}
            onChange={(e) => setLdkServerAddress(e.target.value)}
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="ldk-server-cert">TLS certificate path</Label>
          <Input
            required
            id="ldk-server-cert"
            value={ldkServerTlsCertFile}
            onChange={(e) => setLdkServerTlsCertFile(e.target.value)}
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="ldk-server-api-key">API key (hex)</Label>
          <Input
            required
            id="ldk-server-api-key"
            value={ldkServerApiKey}
            onChange={(e) => setLdkServerApiKey(e.target.value)}
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
