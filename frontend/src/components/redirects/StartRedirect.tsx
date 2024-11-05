import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { useInfo } from "src/hooks/useInfo";

export function StartRedirect({ children }: React.PropsWithChildren) {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info || (info.setupCompleted && !info.running)) {
      if (info && !info.albyAccountConnected && info.albyUserIdentifier) {
        navigate("/alby/auth");
      }
      return;
    }

    navigate("/");
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return children;
}
