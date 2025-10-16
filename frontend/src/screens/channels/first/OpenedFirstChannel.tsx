import TickSVG from "public/images/illustrations/tick.svg";
import { Link } from "react-router-dom";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";

export function OpenedFirstChannel() {
  return (
    <div className="flex flex-col items-center justify-center gap-10 p-5 w-full max-w-md">
      <TwoColumnLayoutHeader
        title="Channel Opened"
        description="Your new lightning channel is ready to use."
      />

      <img src={TickSVG} className="w-48" />

      <Link to="/wallet/receive" className="flex w-full justify-center">
        <Button className="flex-1">Receive Your First Payment</Button>
      </Link>
    </div>
  );
}
