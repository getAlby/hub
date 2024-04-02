import React from "react";
import { useSearchParams } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { Card, CardContent } from "src/components/ui/card";
import { Checkbox } from "src/components/ui/checkbox";
import { Input } from "src/components/ui/input";
import { Label } from "src/components/ui/label";
import { LoadingButton } from "src/components/ui/loading-button";
import { useCSRF } from "src/hooks/useCSRF";
import {
  ConnectPeerRequest,
  Node,
  OpenChannelRequest,
  OpenChannelResponse,
} from "src/types";
import { request } from "src/utils/request";

export default function NewCustomChannel() {
  const [loading, setLoading] = React.useState(false);
  const [localAmount, setLocalAmount] = React.useState("");
  const [nodeDetails, setNodeDetails] = React.useState<Node | undefined>();
  const [isPublic, setPublic] = React.useState(false);

  const [searchParams] = useSearchParams();
  const [pubkey, setPubkey] = React.useState(searchParams.get("pubkey") || "");
  const [host, setHost] = React.useState(searchParams.get("host") || "");
  const { data: csrf } = useCSRF();

  const fetchNodeDetails = React.useCallback(async () => {
    if (!pubkey) {
      return;
    }
    try {
      const data = await request<Node>(
        `/api/mempool/lightning/nodes/${pubkey}`
      );
      setNodeDetails(data);
    } catch (error) {
      console.error(error);
      setNodeDetails(undefined);
    }
  }, [pubkey]);

  React.useEffect(() => {
    fetchNodeDetails();
  }, [fetchNodeDetails]);

  const connectPeer = React.useCallback(async () => {
    if (!csrf) {
      throw new Error("csrf not loaded");
    }
    if (!nodeDetails && !host) {
      throw new Error("node details not found");
    }
    const _host = nodeDetails ? nodeDetails.sockets.split(",")[0] : host;
    const [address, port] = _host.split(":");
    if (!address || !port) {
      throw new Error("host not found");
    }
    console.log(`🔌 Peering with ${pubkey}`);
    const connectPeerRequest: ConnectPeerRequest = {
      pubkey,
      address,
      port: +port,
    };
    await request("/api/peers", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(connectPeerRequest),
    });
  }, [csrf, nodeDetails, pubkey, host]);

  async function openChannel() {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      if (
        isPublic &&
        !confirm(
          `Are you sure you want to open a public channel? in most cases a private channel is recommended.`
        )
      ) {
        return;
      }
      if (
        !confirm(
          `Are you sure you want to peer with ${nodeDetails?.alias || pubkey}?`
        )
      ) {
        return;
      }

      setLoading(true);

      await connectPeer();

      if (
        !confirm(
          `Are you sure you want to open a ${localAmount} sat channel to ${
            nodeDetails?.alias || pubkey
          }?`
        )
      ) {
        setLoading(false);
        return;
      }

      console.log(`🎬 Opening channel with ${pubkey}`);

      const openChannelRequest: OpenChannelRequest = {
        pubkey,
        amount: +localAmount,
        public: isPublic,
      };
      const openChannelResponse = await request<OpenChannelResponse>(
        "/api/channels",
        {
          method: "POST",
          headers: {
            "X-CSRF-Token": csrf,
            "Content-Type": "application/json",
          },
          body: JSON.stringify(openChannelRequest),
        }
      );

      if (!openChannelResponse?.fundingTxId) {
        throw new Error("No funding txid in response");
      }

      alert(`🎉 Published tx: ${openChannelResponse.fundingTxId}`);
    } catch (error) {
      console.error(error);
      alert("Something went wrong: " + error);
    } finally {
      setLoading(false);
    }
  }

  const description = nodeDetails?.alias ? (
    <>
      Open a channel with{` `}
      <span style={{ color: `${nodeDetails.color}` }}>⬤</span>
      {` `}
      {`${nodeDetails.alias}(${nodeDetails.active_channel_count} channels)`}
    </>
  ) : (
    "Connect to other nodes on the lightning network"
  );

  return (
    <div>
      <AppHeader title="Open a channel" description={description}></AppHeader>
      <Card className="p-4">
        <CardContent>
          {nodeDetails && (
            <h3 className="font-medium text-2xl">
              <span style={{ color: `${nodeDetails.color}` }}>⬤</span>
              {nodeDetails.alias && (
                <>
                  {nodeDetails.alias} ({nodeDetails.active_channel_count}{" "}
                  channels)
                </>
              )}
            </h3>
          )}
          <div>{pubkey}</div>

          <div className="flex flex-wrap -mx-3 mt-6">
            <div className="w-full px-3 mb-6 md:mb-0">
              <Label
                className="block uppercase tracking-wide text-primary text-xs font-bold mb-2"
                htmlFor="grid-first-name"
              >
                Peer
              </Label>
              <Input
                className="appearance-none block w-full bg-gray-200 text-gray-700 border rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                type="text"
                value={pubkey}
                placeholder="Pubkey of the peer"
                onChange={(e) => {
                  setPubkey(e.target.value.trim());
                }}
              />
            </div>
            {!nodeDetails && pubkey && (
              <div className="w-full px-3 mb-6 md:mb-0">
                <Label
                  className="block uppercase tracking-wide text-primary text-xs font-bold mb-2"
                  htmlFor="grid-first-name"
                >
                  Host:Port
                </Label>
                <Input
                  className="appearance-none block w-full bg-gray-200 text-gray-700 border rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                  type="text"
                  value={host}
                  placeholder="0.0.0.0:9735"
                  onChange={(e) => {
                    setHost(e.target.value.trim());
                  }}
                />
              </div>
            )}
          </div>

          <div className="flex flex-wrap -mx-3 mt-6">
            <div className="w-full px-3 mb-6 md:mb-0">
              <Label
                className="block uppercase tracking-wide text-primary text-xs font-bold mb-2"
                htmlFor="grid-first-name"
              >
                Amount (sats)
              </Label>
              <Input
                className="appearance-none block w-full bg-gray-200 text-gray-700 border rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
                type="text"
                value={localAmount}
                onChange={(e) => {
                  setLocalAmount(e.target.value.trim());
                }}
              />
            </div>
          </div>

          <div className="flex w-full my-6 items-center">
            <Checkbox
              id="public-channel"
              defaultChecked={isPublic}
              onCheckedChange={() => setPublic(!isPublic)}
              className="mr-2"
            />
            <Label htmlFor="public-channel">Public Channel</Label>
          </div>

          <div className="mt-2">
            <LoadingButton
              size={"sm"}
              disabled={!pubkey || !localAmount || loading}
              onClick={openChannel}
              loading={loading}
            >
              {loading ? "Loading..." : "Open Channel"}
            </LoadingButton>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
