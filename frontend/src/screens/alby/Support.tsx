import {
  LifeBuoy,
  Mail,
  MessageSquare,
  RefreshCw,
  Server,
  Users,
  Zap,
} from "lucide-react";
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
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";

function Support() {
  return (
    <>
      <AppHeader
        title="Support"
        description=""
        contentRight={
          <Dialog>
            <DialogTrigger>
              <Button>✨ Upgrade</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>✨ Upgrade to Alby Pro</DialogTitle>
              </DialogHeader>
              <div className="flex flex-col gap-8">
                <div>
                  <div className="font-semibold mb-2">What you'll get:</div>
                  <ul>
                    <li className="flex flex-row items-center">
                      <RefreshCw className="w-4 h-4 mr-2" />
                      Encrypted remote backups (explain LDK)
                    </li>
                    <li className="flex flex-row items-center">
                      <Users className="w-4 h-4 mr-2" />
                      Unlimited friends & family connections (2-3 free
                      connections)
                    </li>
                    <li className="flex flex-row items-center">
                      <LifeBuoy className="w-4 h-4 mr-2" />
                      Priority Support (pass on param to getalby.com)
                    </li>
                    ----
                    <li className="flex flex-row items-center">
                      <Zap className="w-4 h-4 mr-2" />
                      Customizable lightning address
                    </li>
                    <li className="flex flex-row items-center">
                      <Mail className="w-4 h-4 mr-2" />
                      Email notifications
                    </li>
                    <li className="flex flex-row items-center">
                      <Server className="w-4 h-4 mr-2" />
                      Node monitoring
                    </li>
                  </ul>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <Card className="border border-white">
                    <CardHeader>
                      <CardTitle>Monthly</CardTitle>
                      <CardDescription>3.90 USD / month</CardDescription>
                    </CardHeader>
                  </Card>
                </div>
              </div>
              <DialogFooter>
                <DialogClose asChild>
                  <Button>Upgrade now</Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
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
          buttonText="✨ Upgrade to Alby Pro"
          buttonLink={""}
        />
      </div>
    </>
  );
}

export default Support;
