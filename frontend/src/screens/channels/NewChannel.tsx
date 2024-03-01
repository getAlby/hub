import { Link } from "react-router-dom";
import { useInfo } from "src/hooks/useInfo";

export default function NewChannel() {
  const { data: info } = useInfo();
  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <p>What Type of channel would you like to open?</p>
      {info?.backendType === "LDK" && (
        <Link className="text-purple-500" to="/channels/new/instant">
          Buy Instant Channel
        </Link>
      )}
      <Link className="text-purple-500" to="/channels/new/blocktank">
        Buy Liquidity from Blocktank
      </Link>
      <Link className="text-purple-500" to="/channels/recommended">
        Connect with a Recommended Channel
      </Link>
      <Link className="text-purple-500" to="/channels/new/custom">
        Custom Channel
      </Link>
    </div>
  );
}
