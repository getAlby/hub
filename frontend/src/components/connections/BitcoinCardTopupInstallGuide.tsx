import { useLocation } from "react-router";
import ExternalLink from "src/components/ExternalLink";

// The card topup app supports configuration presets selected via a `provider`
// query param (e.g. linked from the Cards page). Read it straight from the URL
// so the link opens card.albylabs.com pre-configured for the chosen provider.
export function BitcoinCardTopupInstallGuide() {
  const provider = new URLSearchParams(useLocation().search).get("provider");
  const url = provider
    ? `https://card.albylabs.com?provider=${encodeURIComponent(provider)}`
    : "https://card.albylabs.com";

  return (
    <div>
      <ul className="list-inside list-decimal text-muted-foreground">
        <li>
          Open{" "}
          <ExternalLink to={url} className="underline">
            card.albylabs.com
          </ExternalLink>{" "}
          on the device you'll top up from.
        </li>
        <li>
          <span className="font-medium text-foreground">
            Add it to your home screen
          </span>{" "}
          (or bookmark it) so you can reopen it later.
        </li>
        <li>Enter your card's deposit details to set it up.</li>
      </ul>
    </div>
  );
}
