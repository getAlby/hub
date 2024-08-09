import React from "react";
import { useNavigate } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useToast } from "src/components/ui/use-toast";

import { splitSocketAddress } from "src/lib/utils";
import { ConnectPeerRequest } from "src/types";
import { request } from "src/utils/request";

export default function ConnectPeer() {
  const { toast } = useToast();
  const [isLoading, setLoading] = React.useState(false);
  const [connectionString, setConnectionString] = React.useState("");
  const navigate = useNavigate();

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
      toast({ title: "Successfully connected with peer" });
      setConnectionString("");
      navigate("/channels");
    } catch (e) {
      toast({
        variant: "destructive",
        title: "Failed to connect peer: " + e,
      });
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid gap-5">
      <AppHeader
        title="Connect Peer"
        description="Manually connect to a lightning network peer"
      />
      <div className="max-w-lg">
        <form onSubmit={handleSubmit}>
          <div className="">
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
