import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import ExternalLink from "src/components/ExternalLink";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
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
    setupStore.nodeInfo.backendType || "LDK"
  );
  const navigate = useNavigate();

  async function handleSubmit(data: object) {
    setupStore.updateNodeInfo({
      backendType,
      ...data,
    });
    navigate(
      backendType === "BREEZ" ||
        backendType === "GREENLIGHT" ||
        backendType === "LDK"
        ? `/setup/import-mnemonic`
        : `/setup/finish`
    );
  }

  return (
    <>
      <Container>
        <TwoColumnLayoutHeader
          title="Node Setup"
          description="Enter your node connection credentials to connect to your wallet."
        />
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
              <SelectItem value="LND">LND</SelectItem>
              <SelectItem value="PHOENIX">Phoenix</SelectItem>
              <SelectItem value="CASHU">Cashu</SelectItem>
            </SelectContent>
          </Select>
          {backendType === "BREEZ" && <BreezForm handleSubmit={handleSubmit} />}
          {backendType === "GREENLIGHT" && (
            <GreenlightForm handleSubmit={handleSubmit} />
          )}
          {backendType === "LDK" && <LDKForm handleSubmit={handleSubmit} />}
          {backendType === "LND" && <LNDForm handleSubmit={handleSubmit} />}
          {backendType === "PHOENIX" && (
            <PhoenixForm handleSubmit={handleSubmit} />
          )}
          {backendType === "CASHU" && <CashuForm handleSubmit={handleSubmit} />}
        </div>
      </Container>
    </>
  );
}

type SetupFormProps = {
  handleSubmit(data: unknown): void;
};

function CashuForm({ handleSubmit }: SetupFormProps) {
  const [cashuMintUrl, setCashuMintUrl] = React.useState<string>("");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    handleSubmit({ cashuMintUrl });
  }

  return (
    <form className="w-full grid gap-5" onSubmit={onSubmit}>
      <div className="grid gap-1.5">
        <Label
          htmlFor="cashu-mint-url"
          className="flex flex-row justify-between"
        >
          <span>Cashu Mint URL</span>{" "}
          <ExternalLink
            to="https://bitcoinmints.com"
            className="underline hover:no-underline text-xs font-normal"
          >
            Find a mint
          </ExternalLink>
        </Label>
        <Input
          name="cashu-mint-url"
          onChange={(e) => setCashuMintUrl(e.target.value)}
          value={cashuMintUrl}
          id="cashu-mint-url"
          placeholder="https://8333.space:3338"
        />
      </div>

      <Button>Next</Button>
    </form>
  );
}

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

function PhoenixForm({ handleSubmit }: SetupFormProps) {
  const { toast } = useToast();
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

  return (
    <form className="w-full grid gap-5" onSubmit={onSubmit}>
      <div className="grid gap-1.5">
        <Label htmlFor="phoenix-address">Phoneixd Address</Label>
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
        <Input
          name="phoenix-authorization"
          onChange={(e) => setPhoenixdAuthorization(e.target.value)}
          value={phoenixdAuthorization}
          type="password"
          id="phoenix-authorization"
        />
      </div>
      <Button>Next</Button>
    </form>
  );
}
