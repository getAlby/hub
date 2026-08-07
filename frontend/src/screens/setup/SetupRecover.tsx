import { FileKey2Icon, KeyRoundIcon } from "lucide-react";
import { useEffect } from "react";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { LinkButton } from "src/components/ui/custom/link-button";
import useSetupStore from "src/state/SetupStore";

/**
 * Two-button recovery chooser (all users, technical and non-technical):
 * 1) Hub backup (.bkp) — apps + settings + keys
 * 2) Recovery phrase — Lightning access while Greenlight is up; re-pair apps
 */
export function SetupRecover() {
  useEffect(() => {
    useSetupStore.getState().setHasImportedMnemonic(false);
  }, []);

  return (
    <Container>
      <div className="grid gap-5">
        <TwoColumnLayoutHeader
          title="Restore your Hub"
          pageTitle="Restore your Hub"
          description="Choose how you want to get back in. Same options for everyone."
        />
        <div className="flex flex-col gap-3">
          <LinkButton to="/setup/node-restore" className="w-full h-auto py-4">
            <span className="flex flex-col items-start gap-1 text-left w-full">
              <span className="flex items-center gap-2 font-medium">
                <FileKey2Icon className="size-4 shrink-0" />
                Restore Hub backup
              </span>
              <span className="text-xs font-normal opacity-90 pl-6">
                Use your .bkp file. Restores apps, settings, and keys. Best if
                you moved devices.
              </span>
            </span>
          </LinkButton>
          <LinkButton
            to="/setup/password?wallet=import&node=greenlight"
            variant="secondary"
            className="w-full h-auto py-4"
          >
            <span className="flex flex-col items-start gap-1 text-left w-full">
              <span className="flex items-center gap-2 font-medium">
                <KeyRoundIcon className="size-4 shrink-0" />
                I have my recovery phrase
              </span>
              <span className="text-xs font-normal opacity-90 pl-6">
                12 words restore your Greenlight node while Greenlight is
                online. App connections need to be re-created.
              </span>
            </span>
          </LinkButton>
          <LinkButton
            to="/setup/advanced"
            variant="ghost"
            className="w-full text-muted-foreground"
          >
            More options
          </LinkButton>
        </div>
      </div>
    </Container>
  );
}
