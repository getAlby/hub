import React from "react";
import { useLocation } from "react-router-dom";
import AuthCodeForm from "src/components/AuthCodeForm";

import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";

export default function AlbyAuthRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const queryParams = new URLSearchParams(location.search);
  const forceLogin = !!queryParams.get("force_login");
  const url = info?.albyAuthUrl
    ? `${info.albyAuthUrl}${forceLogin ? "&force_login=true" : ""}`
    : undefined;

  React.useEffect(() => {
    if (!info || !url) {
      return;
    }
    if (info.oauthRedirect) {
      window.location.href = url;
    }
  }, [info, url]);

  return !info || info.oauthRedirect || !url ? (
    <Loading />
  ) : (
    <AuthCodeForm url={url} />
  );
}
