import { useInfo } from "src/hooks/useInfo";
import { useLocation, useNavigate } from "react-router-dom";
import React from "react";
import Loading from "src/components/Loading";
import { localStorageKeys } from "src/constants";

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
      if (!info.hasChannels) {
        to = "/channels/first";
      } else if (info.unlocked) {
        const returnTo = window.localStorage.getItem(localStorageKeys.returnTo);
        to = returnTo || "/apps";
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
