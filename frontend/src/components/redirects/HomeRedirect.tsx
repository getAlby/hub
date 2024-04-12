import React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Loading from "src/components/Loading";
import { localStorageKeys } from "src/constants";
import { useInfo } from "src/hooks/useInfo";

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
        if (info.albyAccountConnected) {
          const onboardingSkipped = window.localStorage.getItem(
            localStorageKeys.onboardingSkipped
          );
          if (info.onboardingCompleted || onboardingSkipped) {
            const returnTo = window.localStorage.getItem(
              localStorageKeys.returnTo
            );
            // setTimeout hack needed for React strict mode (in development)
            // because the effect runs twice before the navigation occurs
            setTimeout(() => {
              window.localStorage.removeItem(localStorageKeys.returnTo);
            }, 100);
            to = returnTo || "/wallet";
          } else {
            to = "/onboarding/lightning/migrate-alby";
          }
        } else {
          to = "/alby/auth";
        }
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
