import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Alert from "src/components/Alert";
import ConnectButton from "src/components/ConnectButton";
import Container from "src/components/Container";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import useSetupStore from "src/state/SetupStore";
import { BackendType } from "src/types";
import { handleRequestError } from "src/utils/handleRequestError";
import { request } from "src/utils/request"; // build the project for this to appear

export function SetupNode() {
  const [backendType, setBackendType] = React.useState<BackendType>("BREEZ");
  const { unlockPassword } = useSetupStore();
  const [isConnecting, setConnecting] = React.useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const params = new URLSearchParams(location.search);
  const isNew = params.get("wallet") === "new";

  const { data: info, mutate: refetchInfo } = useInfo();
  const { data: csrf } = useCSRF();

  async function handleSubmit(data: object) {
    if (backendType === "BREEZ") {
      navigate(`/setup/mnemonic${isNew ? "?wallet=new" : ""}`, {
        state: {
          backendType,
          unlockPassword,
          ...data,
        },
      });
      return;
    }

    try {
      setConnecting(true);
      if (!csrf) {
        throw new Error("info not loaded");
      }

      await request("/api/setup", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          backendType,
          unlockPassword,
          ...data,
        }),
      });

      await refetchInfo();
      navigate("/");
    } catch (error) {
      handleRequestError("Failed to connect", error);
    } finally {
      setConnecting(false);
    }
  }

  return (
    <>
      <Container>
        <p className="text-center font-light text-md leading-relaxed dark:text-neutral-400 px-4 mb-4">
          Enter your node connection credentials to connect to your wallet.
        </p>
        {info?.setupCompleted && (
          <Alert type="warn">
            ⚠️ Your node is already setup! only continue if you actually want to
            change your connection settings.
          </Alert>
        )}
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
            disabled={isNew}
            onChange={(e) => setBackendType(e.target.value as BackendType)}
            id="backend-type"
            className="dark:bg-surface-00dp mb-4 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
          >
            <option value={"BREEZ"}>Breez</option>
            <option value={"LND"}>LND</option>
          </select>
          {backendType === "BREEZ" && (
            <BreezForm
              handleSubmit={handleSubmit}
              isConnecting={isConnecting}
            />
          )}
          {backendType === "LND" && (
            <LNDForm handleSubmit={handleSubmit} isConnecting={isConnecting} />
          )}
        </div>
      </Container>
    </>
  );
}

type SetupFormProps = {
  isConnecting: boolean;
  handleSubmit(data: unknown): void;
};

function BreezForm({ isConnecting, handleSubmit }: SetupFormProps) {
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>("");
  const [breezApiKey, setBreezApiKey] = React.useState<string>("");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    // Isn't breezApiKey not required?
    if (!greenlightInviteCode) {
      alert("Please fill out all fields");
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
        <input
          name="greenlight-invite-code"
          onChange={(e) => setGreenlightInviteCode(e.target.value)}
          value={greenlightInviteCode}
          type="password"
          id="greenlight-invite-code"
          autoComplete="new-password"
          placeholder="XXXX-YYYY"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
        <label
          htmlFor="breez-api-key"
          className="block mt-4 mb-2 text-md dark:text-white"
        >
          Breez API Key
        </label>
        <input
          name="breez-api-key"
          onChange={(e) => setBreezApiKey(e.target.value)}
          value={breezApiKey}
          type="password"
          id="breez-api-key"
          autoComplete="new-password"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
      </>
      <ConnectButton
        isConnecting={isConnecting}
        submitText="Next"
        loadingText="Saving..."
      />
    </form>
  );
}

function LNDForm({ isConnecting, handleSubmit }: SetupFormProps) {
  const [lndAddress, setLndAddress] = React.useState<string>("");
  const [lndCertHex, setLndCertHex] = React.useState<string>("");
  const [lndMacaroonHex, setLndMacaroonHex] = React.useState<string>("");
  // TODO: proper onboarding

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!lndAddress || !lndCertHex || !lndMacaroonHex) {
      alert("please fill out all fields");
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
        <input
          name="lnd-address"
          onChange={(e) => setLndAddress(e.target.value)}
          value={lndAddress}
          id="lnd-address"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />

        <label
          htmlFor="lnd-cert-hex"
          className="block mt-4 mb-2 text-md text-gray-900 dark:text-white"
        >
          TLS Certificate (Hex)
        </label>
        <input
          name="lnd-cert-hex"
          onChange={(e) => setLndCertHex(e.target.value)}
          value={lndCertHex}
          type="password"
          id="lnd-cert-hex"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
        <label
          htmlFor="lnd-macaroon-hex"
          className="block mt-4 mb-2 text-md text-gray-900 dark:text-white"
        >
          Admin Macaroon (Hex)
        </label>
        <input
          name="lnd-macaroon-hex"
          onChange={(e) => setLndMacaroonHex(e.target.value)}
          value={lndMacaroonHex}
          type="password"
          id="lnd-macaroon-hex"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
      </>
      <ConnectButton
        isConnecting={isConnecting}
        submitText="Finish"
        loadingText="Saving..."
      />
    </form>
  );
}
