import React from "react";
import { useLocation, useNavigate } from "react-router-dom";

import { useUser } from "@hooks/useUser";

const RETURN_TO_KEY = "returnTo";

export function useLogin() {
  const { data: user, isLoading } = useUser();
  const location = useLocation();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!isLoading) {
      if (user) {
        const returnTo = window.localStorage.getItem(RETURN_TO_KEY);
        if (returnTo) {
          window.localStorage.removeItem(RETURN_TO_KEY);
          navigate(returnTo);
        }
      }
      if (!user && ["/login", "/about"].indexOf(location.pathname) < 0) {
        const returnTo = location.pathname + location.search;
        window.localStorage.setItem(RETURN_TO_KEY, returnTo);

        navigate("/login");
      }
    }
  }, [user, isLoading, location.pathname, location.search, navigate]);

  return null;
}
