import AppHeader from "src/components/AppHeader";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { UpgradeDialog } from "src/components/UpgradeDialog";

function Support() {
  return (
    <>
      <AppHeader
        title="Support"
        description=""
        contentRight={
          <UpgradeDialog>
            <Button variant="premium">Upgrade</Button>
          </UpgradeDialog>
        }
      />
      <div className="flex flex-col gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Where to find help</CardTitle>
          </CardHeader>
          <CardContent>TBD</CardContent>
        </Card>
      </div>
    </>
  );
}

export default Support;
