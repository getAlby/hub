import React from "react";
import AuthCodeForm from "src/components/AuthCodeForm";

import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";

export default function AlbyAuthRedirect() {
  const { data: info } = useInfo();

  React.useEffect(() => {
    if (!info) {
      return;
    }
    if (info.oauthRedirect) {
      window.location.href = info.albyAuthUrl;
    }
  }, [info]);

  return !info || info.oauthRedirect ? <Loading /> : <AuthCodeForm />;
}
