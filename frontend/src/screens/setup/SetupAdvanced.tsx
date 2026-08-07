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
          description="Restore a backup, import a recovery phrase, or pick a custom node backend."
        />
        <div className="flex flex-col gap-3">
          <LinkButton to="/setup/recover" className="w-full">
            Restore Hub or recovery phrase
          </LinkButton>
          <LinkButton
            to="/setup/node-restore"
            variant="secondary"
            className="w-full"
          >
            Import wallet from migration file
          </LinkButton>
          <LinkButton
            to="/setup/password?wallet=import"
            variant="secondary"
            className="w-full"
          >
            Import recovery phrase (choose node)
          </LinkButton>
          <LinkButton
            to="/setup/password"
            variant="secondary"
            className="w-full"
          >
            Create wallet with custom node
          </LinkButton>
        </div>
      </div>
    </Container>
  );
}
