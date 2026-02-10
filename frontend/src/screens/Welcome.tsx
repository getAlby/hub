import React from "react";
import { useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
import { localStorageKeys } from "src/constants";
import { useInfo } from "src/hooks/useInfo";

export function Welcome() {
  const { data: info } = useInfo();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info?.setupCompleted) {
      return;
    }
    navigate("/");
  }, [info, navigate]);

  function navigateToAuthPage(returnTo: string) {
    if (info?.albyAccountConnected) {
      // in case user goes back after authenticating in setup
      // we don't want to show the auth screen twice
      navigate(returnTo);
      return;
    }

    window.localStorage.setItem(localStorageKeys.setupReturnTo, returnTo);

    // by default, allow the user to choose whether or not to connect to alby account
    let navigateTo = "/setup/alby";
    if (info?.oauthRedirect) {
      // if using a custom OAuth client (e.g. Alby Cloud) the user must connect their Alby account
      // but they are already logged in at getalby.com, so it should be an instant redirect.
      navigateTo = "/alby/auth";
    }
    navigate(navigateTo);
  }

  return (
    <Container>
      <title>Welcome - Alby Hub</title>
      <div className="grid text-center gap-5">
        <div className="grid gap-2">
          <h1 className="font-semibold text-2xl font-headline">
            Welcome to Alby Hub
          </h1>
          <p className="text-muted-foreground">
            A powerful, all-in-one bitcoin lightning wallet with the superpower
            of connecting to applications.
          </p>
        </div>
        <div className="grid gap-2">
          <Button
            className="w-full"
            onClick={() =>
              navigateToAuthPage(
                info?.backendType
                  ? "/setup/password?node=preset" // node already setup through env variables
                  : "/setup/password?node=ldk"
              )
            }
          >
            Get Started
            {info?.backendType && ` (${info?.backendType})`}
          </Button>

          {info?.enableAdvancedSetup && (
            <Button
              variant="secondary"
              className="w-full"
              onClick={() => navigateToAuthPage("/setup/advanced")}
            >
              Advanced Setup
            </Button>
          )}
        </div>
      </div>
    </Container>
  );
}
