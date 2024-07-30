import { Boostagram } from "src/types";

function PodcastingInfo({ boost }: { boost: Boostagram }) {
  return (
    <>
      {boost?.message && (
        <div className="mt-6">
          <p>Message</p>
          <p className="text-muted-foreground break-all">{boost.message}</p>
        </div>
      )}
      {boost?.podcast && (
        <div className="mt-6">
          <p>Podcast</p>
          <p className="text-muted-foreground break-all">{boost.podcast}</p>
        </div>
      )}
      {boost?.episode && (
        <div className="mt-6">
          <p>Episode</p>
          <p className="text-muted-foreground break-all">{boost.episode}</p>
        </div>
      )}
      {boost?.action && (
        <div className="mt-6">
          <p>Action</p>
          <p className="text-muted-foreground break-all">{boost.action}</p>
        </div>
      )}
      {boost?.ts && (
        <div className="mt-6">
          <p>Timestamp</p>
          <p className="text-muted-foreground break-all">{boost.ts}</p>
        </div>
      )}
      {boost?.value_msat_total && (
        <div className="mt-6">
          <p>Total amount</p>
          <p className="text-muted-foreground break-all">
            {new Intl.NumberFormat(undefined, {}).format(
              Math.floor(boost.value_msat_total / 1000)
            )}{" "}
            {Math.floor(boost.value_msat_total / 1000) == 1 ? "sat" : "sats"}
          </p>
        </div>
      )}
      {boost?.sender_name && (
        <div className="mt-6">
          <p>Sender</p>
          <p className="text-muted-foreground break-all">{boost.sender_name}</p>
        </div>
      )}
      {boost?.app_name && (
        <div className="mt-6">
          <p>App</p>
          <p className="text-muted-foreground break-all">{boost.app_name}</p>
        </div>
      )}
    </>
  );
}

export default PodcastingInfo;
