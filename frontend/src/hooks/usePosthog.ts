import posthog from "posthog-js";
import React from "react";
import { useInfo } from "src/hooks/useInfo";

export function usePosthog() {
  const { data: info } = useInfo();
  const albyUserIdentifier = info?.albyUserIdentifier;
  const isHttpMode = window.location.protocol.startsWith("http");

  React.useEffect(() => {
    if (!isHttpMode || !albyUserIdentifier) {
      return;
    }
    console.log("Posthog enabled");
    posthog.init("phc_W6d0RRrgfXiYX0pcFBdQHp4mC8HWgUdKQpDZkJYEAiD", {
      api_host: "ph.albylabs.com",
      secure_cookie: true,
      persistence: "cookie",
      loaded: (p) => {
        p.identify(albyUserIdentifier);
        console.log("Posthog loaded");
      },
    });
  }, [albyUserIdentifier, isHttpMode]);
}
