import TickSVG from "public/images/illustrations/tick.svg";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LinkButton } from "src/components/ui/custom/link-button";

type OpenedFirstChannelProps = {
  ctaLabel?: string;
  ctaTo?: string;
};

export function OpenedFirstChannel({
  ctaLabel = "Receive Your First Payment",
  ctaTo = "/wallet/receive",
}: OpenedFirstChannelProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-10 p-5 w-full max-w-md">
      <TwoColumnLayoutHeader
        title="Channel Opened"
        pageTitle="Channel Opened"
        description="Your new lightning channel is ready to use."
      />

      <img src={TickSVG} className="w-48" />

      <LinkButton to={ctaTo} className="flex w-full justify-center">
        {ctaLabel}
      </LinkButton>
    </div>
  );
}
