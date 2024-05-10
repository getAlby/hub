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
            A powerful, all-in-one bitcoin lightning wallet with the superpower
            of connecting to applications.
          </p>
        </div>
        <div className="grid gap-2">
          <Link to="/setup/password" className="w-full">
            <Button className="w-full">Create New Alby Hub</Button>
          </Link>
          {!info?.backendType && (
            <Link to="/setup/password?wallet=import" className="w-full">
              <Button variant="ghost" className="w-full">
                Import Existing Wallet
              </Button>
            </Link>
          )}
        </div>
        <div className="text-sm text-muted-foreground">
          By continuing, you agree to our <br />
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
