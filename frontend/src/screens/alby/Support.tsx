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
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "src/components/ui/alert-dialog";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

function Support() {
  return (
    <>
      <AppHeader
        title="Support"
        description="test"
        contentRight={
          <AlertDialog>
            <AlertDialogTrigger>
              <Button>✨ Upgrade</Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>✨ Upgrade to Alby Pro</AlertDialogTitle>
              </AlertDialogHeader>
              <div className="flex flex-col gap-8">
                <div>
                  <div className="font-semibold mb-2">What you'll get:</div>
                  <ul>
                    <li className="flex flex-row items-center">
                      <RefreshCw className="w-4 h-4 mr-2" />
                      Encrypted remote backups
                    </li>
                    <li className="flex flex-row items-center">
                      <Zap className="w-4 h-4 mr-2" />
                      Customizable lightning address
                    </li>
                    <li className="flex flex-row items-center">
                      <Users className="w-4 h-4 mr-2" />
                      Unlimited friends & family connections
                    </li>
                    <li className="flex flex-row items-center">
                      <Mail className="w-4 h-4 mr-2" />
                      Email notifications
                    </li>
                    <li className="flex flex-row items-center">
                      <Server className="w-4 h-4 mr-2" />
                      Node monitoring
                    </li>
                    <li className="flex flex-row items-center">
                      <LifeBuoy className="w-4 h-4 mr-2" />
                      Priority Support
                    </li>
                  </ul>
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <Card className="border border-white">
                    <CardHeader>
                      <CardTitle>Monthly</CardTitle>
                      <CardDescription>5000 sats / month</CardDescription>
                    </CardHeader>
                  </Card>
                  <Card>
                    <CardHeader>
                      <CardTitle>Yearly</CardTitle>
                      <CardDescription>50,000 sats / year</CardDescription>
                    </CardHeader>
                  </Card>
                </div>
              </div>
              <AlertDialogFooter>
                <AlertDialogAction>Upgrade</AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        }
      />
      <div className="flex flex-col gap-6">
        <EmptyState
          icon={MessageSquare}
          title="Need personal assistance?"
          description="Our team of highly skilled support professionals is happy to lend you a helping hand."
          buttonText="✨ Upgrade to Alby Pro"
          buttonLink={""}
        ></EmptyState>
        <Card>
          <CardHeader>
            <CardTitle>title</CardTitle>
            <CardDescription>lakjsdflk</CardDescription>
          </CardHeader>
          <CardContent>asdf</CardContent>
        </Card>
        <div></div>
      </div>
    </>
  );
}

export default Support;
