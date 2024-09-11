import { Link, useNavigate } from "react-router-dom";

import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Button } from "src/components/ui/button";
import useSetupStore from "src/state/SetupStore";

export function SetupAdvanced() {
  const navigate = useNavigate();
  function importLDKMnemonic() {
    useSetupStore.getState().updateNodeInfo({
      backendType: "LDK",
    });
    navigate("/setup/password?wallet=import");
  }

  return (
    <>
      <Container>
        <div className="grid gap-5">
          <TwoColumnLayoutHeader
            title="Advanced Setup"
            description="Import your Alby Hub, import existing mnemonic or Master Key or create custom wallet."
          />
          <div className="flex flex-col gap-3">
            <Link to="/setup/node-restore">
              <Button className="w-full">Import Wallet with Backup File</Button>
            </Link>
            <Button
              className="w-full"
              variant="secondary"
              onClick={importLDKMnemonic}
            >
              Import Existing Mnemonic
            </Button>
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
