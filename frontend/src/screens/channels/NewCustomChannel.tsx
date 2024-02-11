import React from "react";
import { useSearchParams } from "react-router-dom";
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

  const [searchParams] = useSearchParams();
  const [pubkey, setPubkey] = React.useState(searchParams.get("pubkey") || "");
  const { data: csrf } = useCSRF();

  const fetchNodeDetails = React.useCallback(async () => {
    if (!pubkey) {
      return;
    }
    const response = await fetch(
      `https://mempool.space/api/v1/lightning/nodes/${pubkey}`
    );
    const data = await response.json();
    setNodeDetails(data);
  }, [pubkey]);

  React.useEffect(() => {
    fetchNodeDetails();
  }, [fetchNodeDetails]);

  const connectPeer = React.useCallback(async () => {
    if (!csrf) {
      throw new Error("csrf not loaded");
    }
    if (!nodeDetails) {
      throw new Error("node details not found");
    }
    const host = nodeDetails.sockets.split(",")[0];
    const [address, port] = host.split(":");
    if (!address || !port) {
      throw new Error("host not found");
    }
    console.log(`🔌 Peering with ${pubkey}`);
    const connectPeerRequest: ConnectPeerRequest = {
      pubkey,
      address,
      port: +port,
    };
    await request("/api/peer", {
      method: "POST",
      headers: {
        "X-CSRF-Token": csrf,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(connectPeerRequest),
    });
  }, [csrf, nodeDetails, pubkey]);

  async function openChannel() {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      if (!nodeDetails) {
        throw new Error("node details not found");
      }
      if (
        !confirm(
          `Are you sure you want to open a ${localAmount} sat channel to ${nodeDetails.alias}?`
        )
      ) {
        return;
      }

      setLoading(true);

      await connectPeer();

      console.log(`🎬 Opening channel with ${pubkey}`);

      const openChannelRequest: OpenChannelRequest = {
        pubkey,
        amount: +localAmount,
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

  return (
    <div>
      <h2 className="font-bold text-white text-2xl mt-5">Open a channel</h2>

      <p className="text-gray-500 mb-5">
        {nodeDetails?.alias && (
          <>
            Open a channel with {` `}
            <span style={{ color: `${nodeDetails.color}` }}>⬤</span>
            {nodeDetails.alias}({nodeDetails.active_channel_count} channels)
          </>
        )}
        {!nodeDetails?.alias &&
          "Connect to other nodes on the lightning network"}
      </p>

      <div className=" bg-white shadow-md sm:rounded-lg p-4">
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
            <label
              className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
              htmlFor="grid-first-name"
            >
              Peer
            </label>
            <input
              className="appearance-none block w-full bg-gray-200 text-gray-700 border rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
              type="text"
              value={pubkey}
              placeholder="Pubkey of the peer"
              onChange={(e) => {
                setPubkey(e.target.value.trim());
              }}
            />
          </div>
        </div>

        <div className="flex flex-wrap -mx-3 mt-6">
          <div className="w-full px-3 mb-6 md:mb-0">
            <label
              className="block uppercase tracking-wide text-gray-700 text-xs font-bold mb-2"
              htmlFor="grid-first-name"
            >
              Amount (sats)
            </label>
            <input
              className="appearance-none block w-full bg-gray-200 text-gray-700 border rounded py-3 px-4 mb-3 leading-tight focus:outline-none focus:bg-white"
              type="text"
              value={localAmount}
              onChange={(e) => {
                setLocalAmount(e.target.value.trim());
              }}
            />
          </div>
        </div>

        <div className="mt-2">
          <button
            className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
            disabled={!pubkey || !localAmount || loading}
            onClick={openChannel}
          >
            {loading && (
              <svg
                aria-hidden="true"
                role="status"
                className="inline mr-2 w-4 h-4 text-gray-200 animate-spin dark:text-gray-600"
                viewBox="0 0 100 101"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z"
                  fill="currentColor"
                />
                <path
                  d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z"
                  fill="#1C64F2"
                />
              </svg>
            )}
            {loading ? "Loading..." : "Open Channel"}
          </button>
        </div>
      </div>
    </div>
  );
}
