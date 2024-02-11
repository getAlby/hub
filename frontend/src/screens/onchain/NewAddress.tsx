import React from "react";
import Loading from "src/components/Loading";
import { localStorageKeys } from "src/constants";
import { useCSRF } from "src/hooks/useCSRF";
import { GetOnchainAddressResponse } from "src/types";
import { request } from "src/utils/request";

export default function NewOnchainAddress() {
  const { data: csrf } = useCSRF();
  const [onchainAddress, setOnchainAddress] = React.useState<string>();
  const [isLoading, setLoading] = React.useState(false);

  const getNewAddress = React.useCallback(async () => {
    if (!csrf) {
      return;
    }
    setLoading(true);
    try {
      const response = await request<GetOnchainAddressResponse>(
        "/api/wallet/new-address",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          //body: JSON.stringify({}),
        }
      );
      if (!response?.address) {
        throw new Error("No address in response");
      }
      localStorage.setItem(localStorageKeys.onchainAddress, response.address);
      setOnchainAddress(response.address);
    } catch (error) {
      alert("Failed to request a new address: " + error);
    } finally {
      setLoading(false);
    }
  }, [csrf]);

  React.useEffect(() => {
    const existingAddress = localStorage.getItem(
      localStorageKeys.onchainAddress
    );
    if (existingAddress) {
      setOnchainAddress(existingAddress);
      return;
    }
    getNewAddress();
  }, [getNewAddress]);

  function confirmGetNewAddress() {
    if (confirm("Do you want a fresh address?")) {
      getNewAddress();
    }
  }

  if (!onchainAddress) {
    return <p>Loading...</p>;
  }

  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <p>You can deposit Bitcoin to your wallet address below:</p>
      <input
        className="w-full font-mono shadow-md"
        value={onchainAddress}
      ></input>
      <p className="italic text-sm">
        Wait for one block confirmation after depositing.
      </p>
      <button
        className="flex mt-8 bg-red-300 rounded-lg p-4"
        onClick={confirmGetNewAddress}
        disabled={isLoading}
      >
        Get a new address {isLoading && <Loading />}
      </button>
    </div>
  );
}
