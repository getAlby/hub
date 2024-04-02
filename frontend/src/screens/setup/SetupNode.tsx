import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { useToast } from "src/components/ui/use-toast";
import useSetupStore from "src/state/SetupStore";
import { BackendType } from "src/types";

export function SetupNode() {
  const setupStore = useSetupStore();
  const [backendType, setBackendType] = React.useState<BackendType>(
    setupStore.nodeInfo.backendType || "BREEZ"
  );
  const navigate = useNavigate();
  const location = useLocation();

  const params = new URLSearchParams(location.search);
  const isNew = params.get("wallet") === "new";

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType,
      ...(isNew && { mnemonic: bip39.generateMnemonic(wordlist, 128) }),
      ...data,
    });
    navigate(
      !isNew &&
        (backendType === "BREEZ" ||
          backendType === "GREENLIGHT" ||
          backendType === "LDK")
        ? `/setup/import-mnemonic`
        : `/setup/finish`
    );
  }

  return (
    <>
      <Container>
        <div className="grid gap-5">
          <div className="grid gap-2 text-center">
            <h1 className="font-semibold text-2xl font-headline">Node Setup</h1>
            <p className="text-muted-foreground">
              Enter your node connection credentials to connect to your wallet.
            </p>
          </div>
        </div>
        <div className="w-full mt-5">
          <Label htmlFor="backend-type">Backend Type</Label>
          <Select
            name="backend-type"
            value={backendType}
            onValueChange={(value) => setBackendType(value as BackendType)}
          >
            <SelectTrigger className="mb-5">
              <SelectValue placeholder="Backend" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="LDK">LDK</SelectItem>
              <SelectItem value="BREEZ">Breez</SelectItem>
              <SelectItem value="GREENLIGHT">Greenlight</SelectItem>
              {!isNew && <SelectItem value="LND">LND</SelectItem>}
            </SelectContent>
          </Select>
          {backendType === "BREEZ" && <BreezForm handleSubmit={handleSubmit} />}
          {backendType === "GREENLIGHT" && (
            <GreenlightForm handleSubmit={handleSubmit} />
          )}
          {backendType === "LDK" && <LDKForm handleSubmit={handleSubmit} />}
          {backendType === "LND" && <LNDForm handleSubmit={handleSubmit} />}
        </div>
      </Container>
    </>
  );
}

type SetupFormProps = {
  handleSubmit(data: unknown): void;
};

function BreezForm({ handleSubmit }: SetupFormProps) {
  const { toast } = useToast();
  const setupStore = useSetupStore();
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>(setupStore.nodeInfo.greenlightInviteCode || "");
  const [breezApiKey, setBreezApiKey] = React.useState<string>(
    setupStore.nodeInfo.breezApiKey || ""
  );

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode || !breezApiKey) {
      toast({
        title: "Please fill out all fields",
        variant: "destructive",
      });
      return;
    }
    handleSubmit({
      greenlightInviteCode,
      breezApiKey,
    });
  }

  return (
    <form className="w-full grid gap-5" onSubmit={onSubmit}>
      <div className="grid gap-1.5">
        <Label htmlFor="greenlight-invite-code">Greenlight Invite Code</Label>
        <Input
          name="greenlight-invite-code"
          onChange={(e) => setGreenlightInviteCode(e.target.value)}
          value={greenlightInviteCode}
          type="text"
          id="greenlight-invite-code"
          placeholder="XXXX-YYYY"
        />
      </div>
      <div className="grid gap-1.5">
        <Label htmlFor="breez-api-key">Breez API Key</Label>
        <Input
          name="breez-api-key"
          onChange={(e) => setBreezApiKey(e.target.value)}
          value={breezApiKey}
          autoComplete="off"
          type="text"
          id="breez-api-key"
        />
      </div>
      <Button type="submit">Next</Button>
    </form>
  );
}

function GreenlightForm({ handleSubmit }: SetupFormProps) {
  const setupStore = useSetupStore();
  const { toast } = useToast();
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>(setupStore.nodeInfo.greenlightInviteCode || "");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode) {
      toast({
        title: "Please fill out all fields",
        variant: "destructive",
      });
      return;
    }
    handleSubmit({
      greenlightInviteCode,
    });
  }

  return (
    <form onSubmit={onSubmit} className="w-full grid gap-5">
      <div className="grid gap-1.5">
        <Label htmlFor="greenlight-invite-code">Greenlight Invite Code</Label>
        <Input
          name="greenlight-invite-code"
          onChange={(e) => setGreenlightInviteCode(e.target.value)}
          value={greenlightInviteCode}
          type="text"
          id="greenlight-invite-code"
          placeholder="XXXX-YYYY"
        />
      </div>
      <Button>Next</Button>
    </form>
  );
}

function LDKForm({ handleSubmit }: SetupFormProps) {
  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    handleSubmit({});
  }

  return (
    <form onSubmit={onSubmit} className="w-full grid gap-5">
      <Button>Next</Button>
    </form>
  );
}

function LNDForm({ handleSubmit }: SetupFormProps) {
  const { toast } = useToast();
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

  return (
    <form className="w-full grid gap-5" onSubmit={onSubmit}>
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
  );
}
