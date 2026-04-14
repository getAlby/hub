import React from "react";

import { isHttpMode } from "src/utils/isHttpMode";

export function useRegisterProtocolHandler(basePath: string) {
  React.useEffect(() => {
    if (!isHttpMode() || !("registerProtocolHandler" in navigator)) {
      return;
    }

    try {
      const normalizedBasePath = basePath.replace(/\/$/, "");
      const handlerUrl = `${window.location.origin}${normalizedBasePath}/wallet/send?bip21=%s`;
      navigator.registerProtocolHandler("bitcoin", handlerUrl);
    } catch (e) {
      console.error("Failed to register bitcoin protocol handler", e);
    }
  }, [basePath]);
}
