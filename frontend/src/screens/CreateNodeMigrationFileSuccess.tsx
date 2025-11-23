import { BadgePlusIcon, PowerOffIcon, SaveIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import { Card, CardContent } from "src/components/ui/card";

export function CreateNodeMigrationFileSuccess() {
  return (
    <>
      <div className="p-10">
        <AppHeader
          title="Alby Hub Migration File Saved"
          description="You're ready to move your node to another machine"
          addSidebarTrigger={false}
        />
        <Card>
          <CardContent className="pt-6">
            <div className="flex justify-center flex-col gap-4 text-foreground">
              <div className="flex gap-2 items-center ">
                <div className="shrink-0 ">
                  <SaveIcon className="size-6" />
                </div>
                <span>
                  Your Alby Hub migration file has been successfully saved to
                  your filesystem.
                </span>
              </div>
              <div className="flex gap-2 items-center">
                <div className="shrink-0">
                  <PowerOffIcon className="size-6" />
                </div>
                <span>
                  This Alby Hub has is now in a halted state to prevent further
                  changes.{" "}
                  <b>
                    Do not restart it otherwise your migration file will be
                    invalidated.
                  </b>
                </span>
              </div>
              <div className="flex gap-2 items-center">
                <div className="shrink-0 ">
                  <BadgePlusIcon className="size-6" />
                </div>
                <span>
                  Start your new Alby Hub on another machine and choose the
                  "import backup" option to import your migration file.
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
