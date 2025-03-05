import { MessageSquare } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import EmptyState from "src/components/EmptyState";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
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
            <CardDescription>...</CardDescription>
          </CardHeader>
          <CardContent>
            <ul className="list-disc list-inside">
              <li>Alby Guides</li>
              <li>Discord</li>
              <li>✨ Personal support</li>
            </ul>
          </CardContent>
        </Card>
        <EmptyState
          icon={MessageSquare}
          title="Need personal assistance?"
          description="Our team of highly skilled support professionals is happy to lend you a helping hand."
          buttonText="✨ Upgrade"
          buttonLink={""}
        />
      </div>
    </>
  );
}

export default Support;
