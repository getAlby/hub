import { Link } from "react-router-dom";
import { useInfo } from "src/hooks/useInfo";

export default function NewChannel() {
  const { data: info } = useInfo();
  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <p>What Type of channel would you like to open?</p>
      {info?.backendType === "LDK" && (
        <Link className="text-purple-500" to="instant">
          Buy Instant Channel
        </Link>
      )}
      <Link className="text-purple-500" to="blocktank">
        Buy Liquidity from Blocktank
      </Link>
      <Link className="text-purple-500" to="recommended">
        Connect with a Recommended Channel
      </Link>
      <Link className="text-purple-500" to="custom">
        Custom Channel
      </Link>
    </div>
  );
}
