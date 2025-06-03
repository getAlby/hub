import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";

export function OpenedFirstChannel() {
  return (
    <div className="flex flex-col justify-center gap-5 p-5 max-w-md items-stretch">
      <TwoColumnLayoutHeader
        title="Channel Opened"
        description="Your new lightning channel is ready to use"
      />

      <p>
        Congratulations! Your first lightning channel is active and can be used
        to send and receive payments.
      </p>
      <p>
        To ensure you can both send and receive, make sure to balance your{" "}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/alby-hub/node"
          className="underline"
        >
          channel's liquidity
        </ExternalLink>
        .
      </p>

      <Link to="/wallet" className="flex justify-center mt-8">
        <Button>Go To Your Wallet</Button>
      </Link>
    </div>
  );
}
