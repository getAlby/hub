import { useInfo } from "src/hooks/useInfo";
import { useLocation, useNavigate } from "react-router-dom";
import React from "react";
import Loading from "src/components/Loading";

export function HomeRedirect() {
  const { data: info } = useInfo();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info) {
      return;
    }
    let to: string | undefined;
    if (info.setupCompleted && info.running) {
      if (info.unlocked) {
        to = "/apps";
      } else {
        to = "/unlock";
      }
    } else if (info.setupCompleted && !info.running) {
      to = "/start";
    } else {
      to = "/welcome";
    }
    navigate(to);
  }, [info, location, navigate]);

  if (!info) {
    return <Loading />;
  }
}
