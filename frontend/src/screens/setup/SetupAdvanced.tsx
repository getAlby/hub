import { Link } from "react-router-dom";

import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";

export function SetupAdvanced() {
  return (
    <>
      <Container>
        <div className="grid gap-5">
          <TwoColumnLayoutHeader
            title="Advanced Setup"
            description="Import your Alby Hub, import existing recovery phrase or create custom wallet."
          />
          <div className="flex flex-col gap-3">
            <Link to="/setup/node-restore">
              <Button className="w-full">
                Import Wallet from Migration File
              </Button>
            </Link>
            <Link to="/setup/password?wallet=import" className="w-full">
              <Button className="w-full" variant="secondary">
                Import Existing Recovery Phrase
              </Button>
            </Link>
            <Link to="/setup/password" className="w-full">
              <Button className="w-full" variant="secondary">
                Create Wallet with Custom Node
              </Button>
            </Link>
          </div>
        </div>
      </Container>
    </>
  );
}
