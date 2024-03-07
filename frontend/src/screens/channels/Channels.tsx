import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { useChannels } from "src/hooks/useChannels";
import { useInfo } from "src/hooks/useInfo";
import { useOnchainBalance } from "src/hooks/useOnchainBalance";
import { Node } from "src/types";
import { request } from "src/utils/request";

export default function Channels() {
  const { data: channels } = useChannels();
  const { data: onchainBalance } = useOnchainBalance();
  const [nodes, setNodes] = React.useState<Node[]>([]);
  const { data: info } = useInfo();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info || info.running) {
      return;
    }
    navigate("/");
  }, [info, navigate]);

  const loadNodeStats = React.useCallback(async () => {
    if (!channels) {
      return [];
    }
    const nodes = await Promise.all(
      channels?.map(async (channel): Promise<Node | undefined> => {
        try {
          const response = await request<Node>(
            `/api/mempool/lightning/nodes/${channel.remotePubkey}`
          );
          return response;
        } catch (error) {
          console.error(error);
          return undefined;
        }
      })
    );
    setNodes(nodes.filter((node) => !!node) as Node[]);
  }, [channels]);

  React.useEffect(() => {
    loadNodeStats();
  }, [loadNodeStats]);

  const lightningBalance = channels
    ?.map((channel) => channel.localBalance)
    .reduce((a, b) => a + b, 0);

  return (
    <div>
      <div className="grid gap-6 mb-8 md:grid-cols-3 xl:grid-cols-3">
        <div className="min-w-0 rounded-lg shadow-sm overflow-hidden bg-white dark:bg-gray-800">
          <div className="p-4 flex items-center">
            <div className="p-3 rounded-full text-orange-500 dark:text-orange-100 bg-orange-100 dark:bg-orange-500 mr-4">
              <svg fill="currentColor" viewBox="0 0 20 20" className="w-5 h-5">
                <path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z"></path>
              </svg>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-600 dark:text-gray-400">
                Total number of channels
              </div>
              <div className="text-lg font-semibold text-gray-700 dark:text-gray-200">
                {!channels && (
                  <div>
                    <div className="animate-pulse d-inline ">
                      <div className="h-2.5 bg-gray-200 rounded-full dark:bg-gray-700 w-12 my-2"></div>
                    </div>
                  </div>
                )}
                {channels && channels.length}
              </div>
            </div>
          </div>
        </div>
        <div className="min-w-0 rounded-lg shadow-sm overflow-hidden bg-white dark:bg-gray-800">
          <div className="p-4 flex items-center">
            <div className="p-3 rounded-full text-green-500 dark:text-green-100 bg-green-100 dark:bg-green-500 mr-4">
              <svg fill="currentColor" viewBox="0 0 20 20" className="w-5 h-5">
                <path
                  fillRule="evenodd"
                  d="M4 4a2 2 0 00-2 2v4a2 2 0 002 2V6h10a2 2 0 00-2-2H4zm2 6a2 2 0 012-2h8a2 2 0 012 2v4a2 2 0 01-2 2H8a2 2 0 01-2-2v-4zm6 4a2 2 0 100-4 2 2 0 000 4z"
                  clipRule="evenodd"
                ></path>
              </svg>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">
                Onchain balance
              </p>
              <div className="text-lg font-semibold text-gray-700 dark:text-gray-200">
                {!onchainBalance && (
                  <div>
                    <div className="animate-pulse d-inline ">
                      <div className="h-2.5 bg-gray-200 rounded-full dark:bg-gray-700 w-12 my-2"></div>
                    </div>
                  </div>
                )}
                {onchainBalance && (
                  <span>{formatAmount(onchainBalance.sats * 1000)} sats</span>
                )}
              </div>
            </div>
          </div>
        </div>
        <div className="min-w-0 rounded-lg shadow-sm overflow-hidden bg-white dark:bg-gray-800">
          <div className="p-4 flex items-center">
            <div className="p-3 rounded-full text-blue-500 dark:text-blue-100 bg-blue-100 dark:bg-blue-500 mr-4">
              <svg fill="currentColor" viewBox="0 0 20 20" className="w-5 h-5">
                <path d="M3 1a1 1 0 000 2h1.22l.305 1.222a.997.997 0 00.01.042l1.358 5.43-.893.892C3.74 11.846 4.632 14 6.414 14H15a1 1 0 000-2H6.414l1-1H14a1 1 0 00.894-.553l3-6A1 1 0 0017 3H6.28l-.31-1.243A1 1 0 005 1H3zM16 16.5a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0zM6.5 18a1.5 1.5 0 100-3 1.5 1.5 0 000 3z"></path>
              </svg>
            </div>
            <div>
              <div className="text-sm font-medium text-gray-600 dark:text-gray-400">
                Lightning balance
              </div>
              <div className="text-lg font-semibold text-gray-700 dark:text-gray-200">
                {!channels && (
                  <div>
                    <div className="animate-pulse d-inline ">
                      <div className="h-2.5 bg-gray-200 rounded-full dark:bg-gray-700 w-12 my-2"></div>
                    </div>
                  </div>
                )}
                {lightningBalance !== undefined && (
                  <div>{formatAmount(lightningBalance)} sats</div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      <Link
        to={`/channels/new`}
        className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
      >
        Open a channel
      </Link>
      <Link
        to={`/onchain/new-address`}
        className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 mr-2 dark:bg-blue-600 dark:hover:bg-blue-700 focus:outline-none dark:focus:ring-blue-800"
      >
        Onchain address
      </Link>

      <div className="flex flex-col mt-5">
        <div className="overflow-x-auto shadow-md sm:rounded-lg">
          <div className="inline-block min-w-full align-middle">
            <div className="overflow-hidden ">
              <table className="min-w-full divide-y divide-gray-200 table-fixed dark:divide-gray-700">
                <thead className="bg-gray-100 dark:bg-gray-700">
                  <tr>
                    <th
                      scope="col"
                      className="py-3 px-6 text-xs font-medium tracking-wider text-left text-gray-700 uppercase dark:text-gray-400"
                    >
                      Node pubkey
                    </th>
                    <th
                      scope="col"
                      className="py-3 px-6 text-xs font-medium tracking-wider text-left text-gray-700 uppercase dark:text-gray-400"
                    >
                      Capacity
                    </th>
                    <th
                      scope="col"
                      className="w-96 py-3 px-6 text-xs font-medium tracking-wider text-left text-gray-700 uppercase dark:text-gray-400"
                    >
                      Local / Remote
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200 dark:bg-gray-800 dark:divide-gray-700">
                  {!channels && (
                    <tr>
                      <td colSpan={6} className="text-center p-5">
                        <div
                          role="status"
                          className="animate-pulse flex space-between"
                        >
                          <div className="h-2.5 bg-gray-200 rounded-full w-1/3 dark:bg-gray-700 mr-5"></div>
                          <div className="h-2.5 bg-gray-200 rounded-full w-20 dark:bg-gray-700 mr-5"></div>
                          <div className="h-2.5 bg-gray-200 rounded-full w-20 dark:bg-gray-700"></div>
                          <span className="sr-only">Loading...</span>
                        </div>
                      </td>
                    </tr>
                  )}
                  {channels && channels.length > 0 && (
                    <>
                      {channels.map((channel) => {
                        // const localMaxPercentage =
                        //   maxChannelsBalance.local / 100;
                        // const remoteMaxPercentage =
                        //   maxChannelsBalance.remote / 100;
                        const node = nodes.find(
                          (n) => n.public_key === channel.remotePubkey
                        );
                        const alias = node?.alias || "Unknown";
                        const capacity =
                          channel.localBalance + channel.remoteBalance;
                        const localPercentage =
                          (channel.localBalance / capacity) * 100;
                        const remotePercentage =
                          (channel.remoteBalance / capacity) * 100;

                        return (
                          <tr
                            className="hover:bg-gray-100 dark:hover:bg-gray-700"
                            key={channel.id}
                          >
                            <td className="py-4 px-6 text-sm font-medium text-gray-900 whitespace-nowrap dark:text-white">
                              {channel.active ? "🟢" : "🔴"}{" "}
                              <a
                                className="underline"
                                title={channel.remotePubkey}
                                href={`https://amboss.space/node/${channel.remotePubkey}`}
                                target="_blank"
                                rel="noopener noreferer"
                              >
                                {alias} ({channel.remotePubkey.substring(0, 10)}
                                ...)
                              </a>
                            </td>
                            <td className="py-4 px-6 text-sm font-medium text-gray-500 whitespace-nowrap dark:text-white">
                              {formatAmount(capacity)}
                            </td>
                            <td className="py-4 px-6 text-sm font-light text-right whitespace-nowrap dark:text-gray-400">
                              <div className="flex justify-between">
                                <span>
                                  {formatAmount(channel.localBalance)}
                                </span>
                                <span>
                                  {formatAmount(channel.remoteBalance)}
                                </span>
                              </div>

                              <div className="w-full flex justify-center items-center">
                                <div
                                  className="bg-green-500 h-3 rounded-l-lg"
                                  style={{ width: `${localPercentage}%` }}
                                ></div>
                                <div
                                  className="bg-blue-500 h-3 rounded-r-lg"
                                  style={{ width: `${remotePercentage}%` }}
                                ></div>
                              </div>
                            </td>
                          </tr>
                        );
                      })}
                    </>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

const formatAmount = (amount: number, decimals = 1) => {
  amount /= 1000; //msat to sat
  let i = 0;
  for (i; amount >= 1000; i++) {
    amount /= 1000;
  }
  return amount.toFixed(i > 0 ? decimals : 0) + ["", "k", "M", "G"][i];
};
