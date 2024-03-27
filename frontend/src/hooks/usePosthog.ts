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
      api_host: "https://ph.albylabs.com",
      autocapture: false,
      capture_pageview: false,
      persistence: "localStorage+cookie",
      disable_session_recording: true,
      opt_in_site_apps: true,
      secure_cookie: true,
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
