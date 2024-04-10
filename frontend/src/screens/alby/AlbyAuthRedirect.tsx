import React from "react";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";

export default function AlbyAuthRedirect() {
  const { data: info } = useInfo();
  React.useEffect(() => {
    if (!info) {
      return;
    }
    // FIXME: this won't work for Wails
    window.location.href = info.albyAuthUrl;
  }, [info]);

  return <Loading />;
}
