import { useInfo } from "src/hooks/useInfo";
import { useLocation, useNavigate } from "react-router-dom";
import React from "react";
import Loading from "src/components/Loading";

export function StartRedirect({ children }: React.PropsWithChildren) {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info || (info.setupCompleted && !info.running)) {
      return;
    }
    navigate("/");
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }

  return children;
}
