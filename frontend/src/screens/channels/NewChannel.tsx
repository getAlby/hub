import { Link } from "react-router-dom";

export default function NewChannel() {
  return (
    <div className="flex flex-col justify-center items-center gap-4">
      <p>What Type of channel would you like to open?</p>
      <Link className="text-purple-500" to="/channels/new/blocktank">
        Buy Liquidity from Blocktank
      </Link>
      {/* <Link className="text-purple-500" to="/channels/new/recommended">
        Connect with a Recommended Channel
      </Link> */}
      <Link className="text-purple-500" to="/channels/new/custom">
        Custom Channel
      </Link>
    </div>
  );
}
