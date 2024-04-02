import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import { useInfo } from "src/hooks/useInfo";

export default function NewChannel() {
  const { data: info } = useInfo();
  return (
    <div className="flex flex-col gap-8">
      <div className="grid gap-2 text-center">
        <h1 className="text-2xl font-semibold">Open a new channel</h1>
        <p className="text-muted-foreground">
          Choose how you want to obtain a channel.
        </p>
      </div>
      <div className="flex flex-col justify-center items-center gap-4">
        {info?.backendType === "LDK" && (
          <Link to="instant">
            <Button variant="outline">Buy Instant Channel</Button>
          </Link>
        )}
        <Link to="blocktank">
          <Button variant="outline">Buy Liquidity from Blocktank</Button>
        </Link>
        <Link to="recommended">
          <Button variant="outline">Connect with a Recommended Channel</Button>
        </Link>
        <Link to="custom">
          <Button variant="outline">Custom Channel</Button>
        </Link>
      </div>
    </div>
  );
}
