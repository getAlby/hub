import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import AppHeader from "src/components/AppHeader";
import { LoadingButton } from "src/components/ui/custom/loading-button";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";

import { splitSocketAddress } from "src/lib/utils";
import { ConnectPeerRequest } from "src/types";
import { request } from "src/utils/request";

export default function ConnectPeer() {
  const navigate = useNavigate();
  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);
  const [isLoading, setLoading] = React.useState(false);
  const [connectionString, setConnectionString] = React.useState(
    queryParams.get("peer") ?? ""
  );
  const returnTo = queryParams.get("return_to") ?? "";

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!connectionString) {
      throw new Error("connection details missing");
    }
    try {
      const [pubkey, socketAddress] = connectionString.split("@");
      const { address, port } = splitSocketAddress(socketAddress);
      if (!pubkey || !address || !port) {
        throw new Error("connection details missing");
      }
      console.info(`🔌 Peering with ${pubkey}`);
      const connectPeerRequest: ConnectPeerRequest = {
        pubkey,
        address,
        port: +port,
      };

      setLoading(true);
      await request("/api/peers", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(connectPeerRequest),
      });
      toast("Successfully connected with peer");
      if (returnTo) {
        window.location.href = returnTo;
        return;
      }
      setConnectionString("");
      navigate("/peers");
    } catch (e) {
      toast.error("Failed to connect peer", {
        description: "" + e,
      });
      console.error(e);
    }
    setLoading(false);
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Connect Peer"
        description="Manually connect to a lightning network peer"
      />
      <div className="max-w-lg">
        <form onSubmit={handleSubmit}>
          <div className="grid gap-2">
            <Label htmlFor="connectionString">Peer</Label>
            <Input
              id="connectionString"
              type="text"
              value={connectionString}
              placeholder="pubkey@host:port"
              onChange={(e) => {
                setConnectionString(e.target.value.trim());
              }}
            />
          </div>
          {returnTo && (
            <p className="text-xs text-muted-foreground mt-4">
              You will automatically return to {returnTo}
            </p>
          )}
          <div className="mt-4">
            <LoadingButton
              loading={isLoading}
              type="submit"
              disabled={!connectionString}
              size="lg"
            >
              Connect
            </LoadingButton>
          </div>
        </form>
      </div>
    </div>
  );
}
