import posthog from "posthog-js";
import React from "react";
import { useInfo } from "src/hooks/useInfo";

export function usePosthog() {
  const { data: info } = useInfo();
  const albyUserIdentifier = info?.albyUserIdentifier;
  const isHttpMode = info?.appMode === "HTTP";

  React.useEffect(() => {
    if (!isHttpMode || !albyUserIdentifier) {
      return;
    }
    console.log("Posthog enabled");
    posthog.init("TODO", {
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
