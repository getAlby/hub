import confetti from "canvas-confetti";
import React from "react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";

export function Success() {
  React.useEffect(() => {
    for (let i = 0; i < 10; i++) {
      setTimeout(
        () => {
          confetti({
            origin: {
              x: Math.random(),
              y: Math.random(),
            },
            colors: ["#000", "#333", "#666", "#999", "#BBB", "#FFF"],
          });
        },
        Math.floor(Math.random() * 1000)
      );
    }
  });

  return (
    <div className="flex flex-col justify-center gap-5 p-5 max-w-md items-stretch">
      <TwoColumnLayoutHeader
        title="Channel Opened"
        description="Your new lightning channel is ready to use"
      />

      <p>
        Congratulations! Your channel is active and can be used to send and
        receive payments.
      </p>
      <p>
        To ensure you can both send and receive, make sure to balance your{" "}
        <ExternalLink
          to="https://guides.getalby.com/user-guide/v/alby-account-and-browser-extension/alby-hub/liquidity"
          className="underline"
        >
          channel's liquidity
        </ExternalLink>
        .
      </p>

      <Link to="/" className="flex justify-center mt-8">
        <Button>Go to your wallet</Button>
      </Link>
    </div>
  );
}
