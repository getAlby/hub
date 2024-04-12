import confetti from "canvas-confetti";
import React from "react";
import { Link } from "react-router-dom";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";

export function Success() {
  React.useEffect(() => {
    for (let i = 0; i < 10; i++) {
      setTimeout(() => {
        confetti({
          origin: {
            x: 0.5 + Math.random() * 0.5,
            y: Math.random(),
          },
          colors: ["#000", "#333", "#666", "#999", "#BBB", "#FFF"],
        });
      }, Math.floor(Math.random() * 1000));
    }
  });

  return (
    <div className="flex flex-col justify-center gap-5 p-5 max-w-md items-stretch">
      <TwoColumnLayoutHeader
        title="Setup Completed"
        description="Your Alby Account is now self-sovereign"
      />

      <p>
        Congratulations! You're now running your own lightning node connected to
        the lightning network.
      </p>

      <Link to="/" className="flex justify-center mt-8">
        <Button>Go to your new wallet</Button>
      </Link>
    </div>
  );
}
