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
  const [clnCaCert, setClnCaCert] = React.useState<string>(
    setupStore.nodeInfo.clnCaCert || ""
  );
  const [clnClientCert, setClnClientCert] = React.useState<string>(
    setupStore.nodeInfo.clnClientCert || ""
  );
  const [clnClientKey, setClnClientKey] = React.useState<string>(
    setupStore.nodeInfo.clnClientKey || ""
  );
  const [clnAddressHold, setClnAddressHold] = React.useState<string>(
    setupStore.nodeInfo.clnAddressHold || ""
  );
  const [clnCaCertHold, setClnCaCertHold] = React.useState<string>(
    setupStore.nodeInfo.clnCaCertHold || ""
  );
  const [clnClientCertHold, setClnClientCertHold] = React.useState<string>(
    setupStore.nodeInfo.clnClientCertHold || ""
  );
  const [clnClientKeyHold, setClnClientKeyHold] = React.useState<string>(
    setupStore.nodeInfo.clnClientKeyHold || ""
  );

  // TODO: proper onboarding
  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    handleSubmit({
      clnAddress,
      clnCaCert,
      clnClientCert,
      clnClientKey,
      clnAddressHold,
      clnCaCertHold,
      clnClientCertHold,
      clnClientKeyHold,
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
          <Label htmlFor="cln-ca-cert">CLN CA Cert</Label>
          <Input
            required
            name="cln-ca-cert"
            onChange={(e) => setClnCaCert(e.target.value)}
            value={clnCaCert}
            type="text"
            id="cln-ca-cert"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-client-cert">CLN Client Cert</Label>
          <Input
            required
            name="cln-client-cert"
            onChange={(e) => setClnClientCert(e.target.value)}
            value={clnClientCert}
            type="text"
            id="cln-client-cert"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-client-key">CLN Client Key</Label>
          <Input
            required
            name="cln-client-key"
            onChange={(e) => setClnClientKey(e.target.value)}
            value={clnClientKey}
            type="text"
            id="cln-client-key"
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
        <div className="grid gap-1.5">
          <Label htmlFor="cln-ca-cert-hold">
            (optional) CLN hold plugin CA Cert
          </Label>
          <Input
            name="cln-ca-cert-hold"
            onChange={(e) => setClnCaCertHold(e.target.value)}
            value={clnCaCertHold}
            type="text"
            id="cln-ca-cert-hold"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-client-cert-hold">
            (optional) CLN hold plugin Client Cert
          </Label>
          <Input
            name="cln-client-cert-hold"
            onChange={(e) => setClnClientCertHold(e.target.value)}
            value={clnClientCertHold}
            type="text"
            id="cln-client-cert-hold"
          />
        </div>
        <div className="grid gap-1.5">
          <Label htmlFor="cln-client-key-hold">
            (optional) CLN hold plugin Client Key
          </Label>
          <Input
            name="cln-client-key-hold"
            onChange={(e) => setClnClientKeyHold(e.target.value)}
            value={clnClientKeyHold}
            type="text"
            id="cln-client-key-hold"
          />
        </div>
        <Button>Next</Button>
      </form>
    </Container>
  );
}
