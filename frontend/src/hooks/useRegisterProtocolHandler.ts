import React from "react";

import { isHttpMode } from "src/utils/isHttpMode";

const STORAGE_KEY = "bitcoin-protocol-handler-registered";

export function useRegisterProtocolHandler(basePath: string) {
  React.useEffect(() => {
    if (!isHttpMode() || !("registerProtocolHandler" in navigator)) {
      return;
    }

    const normalizedBasePath = basePath.replace(/\/$/, "");
    const handlerUrl = `${window.location.origin}${normalizedBasePath}/wallet/send?bip21=%s`;

    // Browsers re-prompt every time registerProtocolHandler is called if the
    // user previously dismissed the prompt without accepting or denying it.
    // Limit to once per browser session so we don't ask on every page load,
    // but users still get re-asked occasionally if they didn't opt in.
    if (sessionStorage.getItem(STORAGE_KEY) === handlerUrl) {
      return;
    }

    try {
      navigator.registerProtocolHandler("bitcoin", handlerUrl);
      sessionStorage.setItem(STORAGE_KEY, handlerUrl);
    } catch (e) {
      console.error("Failed to register bitcoin protocol handler", e);
    }
  }, [basePath]);
}
