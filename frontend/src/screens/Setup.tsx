import React from "react";
import { useNavigate } from "react-router-dom";
import { useCSRF } from "src/hooks/useCSRF";
import { BackendType } from "src/types";
import { request, handleFetchError } from "src/utils/request";

export function Setup() {
  const [backendType, setBackendType] = React.useState<BackendType>("BREEZ");
  const navigate = useNavigate();

  const { data: csrf } = useCSRF();

  async function handleSubmit(data: object) {
    try {
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

      navigate("/apps");
    } catch (error) {
      handleFetchError("Failed to connect", error);
    }
  }

  return (
    <>
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
        className="mb-4 bg-gray-50 border border-gray-300 text-gray-900 focus:ring-purple-700 dark:focus:ring-purple-600 dark:ring-offset-gray-800 focus:ring-2 text-sm rounded-lg block w-full p-2.5 dark:bg-surface-00dp dark:border-gray-700 dark:placeholder-gray-400 dark:text-white"
      >
        <option value={"BREEZ"}>Breez</option>
        <option value={"LND"}>LND</option>
      </select>

      {backendType === "BREEZ" && <BreezForm handleSubmit={handleSubmit} />}
      {backendType === "LND" && <p>Coming soon</p>}
    </>
  );
}

function ConnectButton() {
  return (
    <button
      type="submit"
      className="mt-4 inline-flex w-full bg-purple-700 cursor-pointer dark:text-neutral-200 duration-150 focus-visible:ring-2 focus-visible:ring-offset-2 focus:outline-none font-medium hover:bg-purple-900 items-center justify-center px-5 py-3 rounded-md shadow text-white transition"
    >
      Connect
    </button>
  );
}

type SetupFormProps = {
  handleSubmit(data: unknown): void;
};

function BreezForm({ handleSubmit }: SetupFormProps) {
  const [greenlightInviteCode, setGreenlightInviteCode] =
    React.useState<string>("");
  const [breezMnemonic, setBreezMnemonic] = React.useState<string>("");

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!greenlightInviteCode || !breezMnemonic) {
      alert("please fill out all fields");
      return;
    }
    handleSubmit({
      greenlightInviteCode,
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
          className="bg-gray-50 border border-gray-300 text-gray-900 focus:ring-purple-700 dark:focus:ring-purple-600 dark:ring-offset-gray-800 focus:ring-2 text-sm rounded-lg block w-full p-2.5 dark:bg-surface-00dp dark:border-gray-700 dark:placeholder-gray-400 dark:text-white"
        />
        <label
          htmlFor="greenlight-invite-code"
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
          className="bg-gray-50 border border-gray-300 text-gray-900 focus:ring-purple-700 dark:focus:ring-purple-600 dark:ring-offset-gray-800 focus:ring-2 text-sm rounded-lg block w-full p-2.5 dark:bg-surface-00dp dark:border-gray-700 dark:placeholder-gray-400 dark:text-white"
        />
      </>
      <ConnectButton />
    </form>
  );
}
