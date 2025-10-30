import { FormattedBitcoinAmount } from "src/components/FormattedBitcoinAmount";
import { Boostagram } from "src/types";

function PodcastingInfo({ boost }: { boost: Boostagram }) {
  return (
    <>
      {boost.message && (
        <div className="mt-6">
          <p>Message</p>
          <p className="text-muted-foreground break-all">{boost.message}</p>
        </div>
      )}
      {boost.podcast && (
        <div className="mt-6">
          <p>Podcast</p>
          <p className="text-muted-foreground break-all">{boost.podcast}</p>
        </div>
      )}
      {boost.episode && (
        <div className="mt-6">
          <p>Episode</p>
          <p className="text-muted-foreground break-all">{boost.episode}</p>
        </div>
      )}
      {boost.action && (
        <div className="mt-6">
          <p>Action</p>
          <p className="text-muted-foreground break-all">{boost.action}</p>
        </div>
      )}
      {boost.ts && (
        <div className="mt-6">
          <p>Timestamp</p>
          <p className="text-muted-foreground break-all">{boost.ts}</p>
        </div>
      )}
      {boost.valueMsatTotal && (
        <div className="mt-6">
          <p>Total amount</p>
          <p className="text-muted-foreground break-all sensitive">
            <FormattedBitcoinAmount amount={boost.valueMsatTotal} />
          </p>
        </div>
      )}
      {boost.senderName && (
        <div className="mt-6">
          <p>Sender</p>
          <p className="text-muted-foreground break-all">{boost.senderName}</p>
        </div>
      )}
      {boost.appName && (
        <div className="mt-6">
          <p>App</p>
          <p className="text-muted-foreground break-all">{boost.appName}</p>
        </div>
      )}
    </>
  );
}

export default PodcastingInfo;
