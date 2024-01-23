import React from "react";
import { useNavigate } from "react-router-dom";
import ConnectButton from "src/components/ConnectButton";
import { useCSRF } from "src/hooks/useCSRF";
import { useInfo } from "src/hooks/useInfo";
import { BackendType } from "src/types";
import { request, handleRequestError } from "src/utils/request";

export function Setup() {
  const [backendType, setBackendType] = React.useState<BackendType>("BREEZ");
  const [isConnecting, setConnecting] = React.useState(false);
  const navigate = useNavigate();

  const { data: info } = useInfo();
  const { data: csrf } = useCSRF();

  async function handleSubmit(data: object) {
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
          ...data,
        }),
      });

      navigate("/");
    } catch (error) {
      handleRequestError("Failed to connect", error);
    } finally {
      setConnecting(false);
    }
  }

  return (
    <>
      {info?.setupCompleted && (
        <p className="mb-4 text-red-500">
          Your node is already setup! only continue if you actually want to
          change your connection settings.
        </p>
      )}
      <p className="mb-4">
        Enter your node connection credentials to connect to your wallet.
      </p>
      <label
        htmlFor="backend-type"
        className="block font-medium text-gray-900 dark:text-white"
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
        <option value={"LND"}>LND</option>
      </select>

      {backendType === "BREEZ" && (
        <BreezForm handleSubmit={handleSubmit} isConnecting={isConnecting} />
      )}
      {backendType === "LND" && (
        <LNDForm handleSubmit={handleSubmit} isConnecting={isConnecting} />
      )}
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
  const [breezMnemonic, setBreezMnemonic] = React.useState<string>("");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode || !breezMnemonic) {
      alert("please fill out all fields");
      return;
    }
    handleSubmit({
      greenlightInviteCode,
      breezApiKey,
      breezMnemonic,
    });
  }

  return (
    <form onSubmit={onSubmit}>
      <>
        <label
          htmlFor="greenlight-invite-code"
          className="block font-medium text-gray-900 dark:text-white"
        >
          Greenlight Invite Code
        </label>
        <input
          name="greenlight-invite-code"
          onChange={(e) => setGreenlightInviteCode(e.target.value)}
          value={greenlightInviteCode}
          type="password"
          id="greenlight-invite-code"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
        <label
          htmlFor="breez-api-key"
          className="mt-4 block font-medium text-gray-900 dark:text-white"
        >
          Breez API Key
        </label>
        <input
          name="breez-api-key"
          onChange={(e) => setBreezApiKey(e.target.value)}
          value={breezApiKey}
          type="password"
          id="breez-api-key"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
        <label
          htmlFor="mnemonic"
          className="mt-4 block font-medium text-gray-900 dark:text-white"
        >
          BIP39 Mnemonic
        </label>
        <input
          name="mnemonic"
          onChange={(e) => setBreezMnemonic(e.target.value)}
          value={breezMnemonic}
          type="password"
          id="mnemonic"
          className="dark:bg-surface-00dp block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:ring-2 focus:ring-purple-700 dark:border-gray-700 dark:text-white dark:placeholder-gray-400 dark:ring-offset-gray-800 dark:focus:ring-purple-600"
        />
      </>
      <ConnectButton isConnecting={isConnecting} />
    </form>
  );
}

function LNDForm({ isConnecting, handleSubmit }: SetupFormProps) {
  const [lndAddress, setLndAddress] = React.useState<string>("");
  const [lndCertHex, setLndCertHex] = React.useState<string>("");
  const [lndMacaroonHex, setLndMacaroonHex] = React.useState<string>("");

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
    <form onSubmit={onSubmit}>
      <>
        <label
          htmlFor="lnd-address"
          className="block font-medium text-gray-900 dark:text-white"
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
          className="mt-4 block font-medium text-gray-900 dark:text-white"
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
          className="mt-4 block font-medium text-gray-900 dark:text-white"
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
      <ConnectButton isConnecting={isConnecting} />
    </form>
  );
}
