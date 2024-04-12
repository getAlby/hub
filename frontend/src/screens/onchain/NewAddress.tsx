import React from "react";
import Loading from "src/components/Loading";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
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
    return (
      <div className="flex justify-center">
        <Loading />
      </div>
    );
  }

  return (
    <div className="grid gap-8">
      <TwoColumnLayoutHeader
        title="On Chain Address"
        description="You can deposit Bitcoin to your wallet address below:"
      />
      <div className="grid gap-2">
        <Label htmlFor="text">On Chain Address</Label>
        <Input type="text" value={onchainAddress} />
        <p className="italic text-sm">
          Wait for one block confirmation after depositing.
        </p>
      </div>

      <div>
        <LoadingButton
          size={"sm"}
          onClick={confirmGetNewAddress}
          disabled={isLoading}
          loading={isLoading}
        >
          Get a new address
        </LoadingButton>
      </div>
    </div>
  );
}
