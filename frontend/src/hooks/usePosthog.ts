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
    posthog.init("phc_W6d0RRrgfXiYX0pcFBdQHp4mC8HWgUdKQpDZkJYEAiD", {
      api_host: "ph.albylabs.com",
      secure_cookie: true,
      persistence: "cookie",
      capture_pageview: false,
      autocapture: false,
      session_recording: {
        maskAllInputs: true,
        maskTextSelector: ".sensitive",
      },
      loaded: (p) => {
        p.onFeatureFlags(() => {
          if (p.isFeatureEnabled("rec-session")) {
            p.startSessionRecording();
          }
        });
        p.identify(albyUserIdentifier);
      },
    });
  }, [albyUserIdentifier, isHttpMode]);
}
