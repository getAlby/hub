import { useCSRF } from "src/hooks/useCSRF";
import { useNodeConnectionInfo } from "src/hooks/useNodeConnectionInfo";
import { ConnectPeerRequest } from "src/types";
import { request } from "src/utils/request";

export default function NewBlocktankChannel() {
  const { data: connectionInfo } = useNodeConnectionInfo();
  const { data: csrf } = useCSRF();
  if (!connectionInfo) {
    return <p>Loading...</p>;
  }

  async function peer() {
    try {
      if (!csrf) {
        throw new Error("csrf not loaded");
      }
      const connectPeerRequest: ConnectPeerRequest = {
        pubkey:
          "0296b2db342fcf87ea94d981757fdf4d3e545bd5cef4919f58b5d38dfdd73bf5c9",
        address: "130.211.95.29",
        port: 9735,
      };
      await request("/api/peers", {
        method: "POST",
        headers: {
          "X-CSRF-Token": csrf,
          "Content-Type": "application/json",
        },
        body: JSON.stringify(connectPeerRequest),
      });
      alert("Connected to peer successfully!");
    } catch (error) {
      alert("Failed to connect to peer: " + error);
    }
  }

  const connectionString = connectionInfo.pubkey; //`${connectionInfo.pubkey}@${connectionInfo.address}:${connectionInfo.port}`;

  // TODO: replace with https://github.com/synonymdev/blocktank-client
  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <button className="shadow-lg p-4 bg-green-500 rounded-lg" onClick={peer}>
        Click here to peer with blocktank
      </button>
      <p>
        After <b>connecting to the peer above</b> paying for the liquidity,
        select the <b>claim manually</b> option and paste the connection string
        below, and then click <b>open channel</b>.
      </p>
      <div className="flex w-full">
        <input
          className="flex-1 font-mono shadow-md"
          value={connectionString}
        ></input>
      </div>
      <iframe
        style={{ margin: "auto", minWidth: "480px", minHeight: "800px" }}
        id="widget"
        src="https://widget.synonym.to/"
        seamless
      ></iframe>
    </div>
  );
}
