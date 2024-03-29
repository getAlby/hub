import * as bip39 from "@scure/bip39";
import { wordlist } from "@scure/bip39/wordlists/english";
import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import Input from "src/components/Input";
import toast from "src/components/Toast";
import { Button } from "src/components/ui/button";
import { LoadingButton } from "src/components/ui/loading-button";
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
        <p className="text-center font-light text-md leading-relaxed dark:text-neutral-400 px-4 mb-4">
          Enter your node connection credentials to connect to your wallet.
        </p>
        <div className="w-full mt-4">
          <label
            htmlFor="backend-type"
            className="block mb-2 text-md text-gray-900 dark:text-white"
          >
            Backend Type
          </label>
          <select
            name="backend-type"
            value={backendType}
            onChange={(e) => setBackendType(e.target.value as BackendType)}
            id="backend-type"
            className="dark:bg-surface-00dp mb-4 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
          >
            <option value={"BREEZ"}>Breez</option>
            <option value={"GREENLIGHT"}>Greenlight</option>
            <option value={"LDK"}>LDK</option>
            {!isNew && <option value={"LND"}>LND</option>}
          </select>
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
  const setupStore = useSetupStore();
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>(setupStore.nodeInfo.greenlightInviteCode || "");
  const [breezApiKey, setBreezApiKey] = React.useState<string>(
    setupStore.nodeInfo.breezApiKey || ""
  );

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode || !breezApiKey) {
      toast.error("Please fill out all fields");
      return;
    }
    handleSubmit({
      greenlightInviteCode,
      breezApiKey,
    });
  }

  return (
    <form className="w-full" onSubmit={onSubmit}>
      <>
        <label
          htmlFor="greenlight-invite-code"
          className="block mb-2 text-md dark:text-white"
        >
          Greenlight Invite Code
        </label>
        <Input
          name="greenlight-invite-code"
          onChange={(e) => setGreenlightInviteCode(e.target.value)}
          value={greenlightInviteCode}
          type="text"
          id="greenlight-invite-code"
          placeholder="XXXX-YYYY"
        />
        <label
          htmlFor="breez-api-key"
          className="block mt-4 mb-2 text-md dark:text-white"
        >
          Breez API Key
        </label>
        <Input
          name="breez-api-key"
          onChange={(e) => setBreezApiKey(e.target.value)}
          value={breezApiKey}
          autoComplete="off"
          type="text"
          id="breez-api-key"
        />
      </>
      <LoadingButton type="submit">Next</LoadingButton>
    </form>
  );
}

function GreenlightForm({ handleSubmit }: SetupFormProps) {
  const setupStore = useSetupStore();
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>(setupStore.nodeInfo.greenlightInviteCode || "");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode) {
      toast.error("please fill out all fields");
      return;
    }
    handleSubmit({
      greenlightInviteCode,
    });
  }

  return (
    <form onSubmit={onSubmit} className="w-full">
      <>
        <label
          htmlFor="greenlight-invite-code"
          className="block mb-2 text-md dark:text-white"
        >
          Greenlight Invite Code
        </label>
        <Input
          name="greenlight-invite-code"
          onChange={(e) => setGreenlightInviteCode(e.target.value)}
          value={greenlightInviteCode}
          type="text"
          id="greenlight-invite-code"
          placeholder="XXXX-YYYY"
        />
      </>
      <LoadingButton>Next</LoadingButton>
    </form>
  );
}

function LDKForm({ handleSubmit }: SetupFormProps) {
  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    handleSubmit({});
  }

  return (
    <form onSubmit={onSubmit} className="w-full">
      <Button>Next</Button>
    </form>
  );
}

function LNDForm({ handleSubmit }: SetupFormProps) {
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
      toast.error("please fill out all fields");
      return;
    }
    handleSubmit({
      lndAddress,
      lndCertHex,
      lndMacaroonHex,
    });
  }

  return (
    <form className="w-full" onSubmit={onSubmit}>
      <>
        <label
          htmlFor="lnd-address"
          className="block mb-2 text-md dark:text-white"
        >
          LND Address (GRPC)
        </label>
        <Input
          name="lnd-address"
          onChange={(e) => setLndAddress(e.target.value)}
          value={lndAddress}
          id="lnd-address"
        />

        <label
          htmlFor="lnd-cert-hex"
          className="block mt-4 mb-2 text-md text-gray-900 dark:text-white"
        >
          TLS Certificate (Hex)
        </label>
        <Input
          name="lnd-cert-hex"
          onChange={(e) => setLndCertHex(e.target.value)}
          value={lndCertHex}
          type="text"
          id="lnd-cert-hex"
        />
        <label
          htmlFor="lnd-macaroon-hex"
          className="block mt-4 mb-2 text-md text-gray-900 dark:text-white"
        >
          Admin Macaroon (Hex)
        </label>
        <Input
          name="lnd-macaroon-hex"
          onChange={(e) => setLndMacaroonHex(e.target.value)}
          value={lndMacaroonHex}
          type="text"
          id="lnd-macaroon-hex"
        />
      </>
      <LoadingButton>Next</LoadingButton>
    </form>
  );
}
