import { BadgePlus, PowerOff, Save } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import { Card, CardContent } from "src/components/ui/card";

export function BackupNodeSuccess() {
  return (
    <>
      <div className="p-10">
        <AppHeader
          title="Backup Successful"
          description="You're ready to move your node to another machine"
        />
        <Card>
          <CardContent className="pt-6">
            <div className="flex justify-center flex-col gap-4 text-foreground">
              <div className="flex gap-2 items-center ">
                <div className="shrink-0 ">
                  <Save className="w-6 h-6" />
                </div>
                <span>
                  Your Alby Hub has been successfully backed up and saved to
                  your filesystem.
                </span>
              </div>
              <div className="flex gap-2 items-center">
                <div className="shrink-0">
                  <PowerOff className="w-6 h-6" />
                </div>
                <span>
                  This Alby Hub has is now in a halted state to prevent further
                  changes.{" "}
                  <b>
                    Do not restart it otherwise your backup will be invalidated.
                  </b>
                </span>
              </div>
              <div className="flex gap-2 items-center">
                <div className="shrink-0 ">
                  <BadgePlus className="w-6 h-6" />
                </div>
                <span>
                  Start your new Alby Hub on another machine and choose the
                  "import backup" option to import your backup file.
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
