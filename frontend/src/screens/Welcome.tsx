import React from "react";
import { Link, useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
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

  return (
    <Container>
      <div className="grid text-center gap-5">
        <div className="grid gap-2">
          <h1 className="font-semibold text-2xl font-headline">
            Welcome to Alby Hub
          </h1>
          <p className="text-muted-foreground">
            Connect your lightning wallet to dozens of apps and enjoy seamless
            in-app payments.
          </p>
        </div>
        <Link to="/setup/password" className="w-full">
          <Button size="lg" className="w-full">
            Continue
          </Button>
        </Link>
        {!info?.backendType && (
          <Link to="/setup/password?wallet=import">
            <Button variant="link" size="sm">
              Import Existing Wallet
            </Button>
          </Link>
        )}
        <div className="text-sm text-muted-foreground">
          By clicking "Continue" or "Import Existing Wallet", you agree to our{" "}
          <br />
          <Link to="#" className="underline">
            Terms of Service
          </Link>{" "}
          and{" "}
          <Link to="#" className="underline">
            Privacy Policy
          </Link>
        </div>
      </div>
    </Container>
  );
}
