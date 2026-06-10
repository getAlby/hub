import { useEffect } from "react";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LinkButton } from "src/components/ui/custom/link-button";
import useSetupStore from "src/state/SetupStore";

export function SetupAdvanced() {
  useEffect(() => {
    // in case the user goes back, reset the imported mnemonic flag
    useSetupStore.getState().setHasImportedMnemonic(false);
  }, []);

  return (
    <Container>
      <div className="grid gap-5">
        <TwoColumnLayoutHeader
          title="Advanced Setup"
          pageTitle="Advanced Setup"
          description="Import your Alby Hub, import existing recovery phrase or create custom wallet."
        />
        <div className="flex flex-col gap-3">
          <LinkButton to="/setup/node-restore" className="w-full">
            Import Wallet from Migration File
          </LinkButton>
          <LinkButton
            to="/setup/password?wallet=import"
            variant="secondary"
            className="w-full"
          >
            Import Existing Recovery Phrase
          </LinkButton>
          <LinkButton
            to="/setup/password"
            variant="secondary"
            className="w-full"
          >
            Create Wallet with Custom Node
          </LinkButton>
        </div>
      </div>
    </Container>
  );
}
