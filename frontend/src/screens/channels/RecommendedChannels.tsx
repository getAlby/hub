import { Link } from "react-router-dom";
import { useBalances } from "src/hooks/useBalances";

type RecommendedNode = {
  title: string;
  pubkey: string;
  min: number;
};

const RECOMMENDED_NODES: RecommendedNode[] = [
  {
    title: "ACINQ",
    pubkey:
      "03864ef025fde8fb587d989186ce6a4a186895ee44a926bfc370e2c366597a3f8f",
    min: 550000,
  },
  {
    title: "deezy prime ⚡ ✨",
    pubkey:
      "0214ec5397050f7eec8e574d1d6feaa0c992169ed107056e6bd57aed1e94714851",
    min: 550000,
  },
  {
    title: "bitfinex.com",
    pubkey:
      "03cde60a6323f7122d5178255766e38114b4722ede08f7c9e0c5df9b912cc201d6",
    min: 550000,
  },
  {
    title: "Breez",
    pubkey:
      "031015a7839468a3c266d662d5bb21ea4cea24226936e2864a7ca4f2c3939836e0",
    min: 10100000,
  },
  {
    title: "kappa",
    pubkey:
      "0324ba2392e25bff76abd0b1f7e4b53b5f82aa53fddc3419b051b6c801db9e2247",
    min: 350000,
  },
  {
    title: "Voltage.cloud (C2)",
    pubkey:
      "02cfdc6b60e5931d174a342b20b50d6a2a17c6e4ef8e077ea54069a3541ad50eb0",
    min: 350000,
  },
];

export default function RecommendedChannels() {
  const { data: balances } = useBalances();
  return (
    <div className="flex flex-col gap-2">
      <h1>Recommended Channel Peers</h1>
      {RECOMMENDED_NODES.map((node) => (
        <Link
          key={node.pubkey}
          to={`/channels/new/custom?pubkey=${node.pubkey}`}
        >
          <div className="w-full rounded bg-neutral-100 p-4 flex flex-col">
            <div className="flex justify-between items-center">
              <h2 className="font-bold">{node.title}</h2>
              <div className="flex gap-4">
                <p className="break-all text-sm">{node.pubkey}</p>
              </div>
            </div>
            <p className="text-sm">
              Min channel size:{" "}
              <span
                className={`${
                  node.min <= (balances?.onchain.spendable || 0)
                    ? "text-green-500"
                    : "text-red-500"
                }`}
              >
                {node.min}
              </span>
            </p>
          </div>
        </Link>
      ))}
    </div>
  );
}
