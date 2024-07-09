import {
  CheckCircle2,
  CircleX,
  EditIcon,
  ExternalLinkIcon,
  Link2Icon,
  ZapIcon,
} from "lucide-react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import UserAvatar from "src/components/UserAvatar";
import { AppCardConnectionInfo } from "src/components/connections/AppCardConnectionInfo";
import { AppCardNotice } from "src/components/connections/AppCardNotice";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
import { Separator } from "src/components/ui/separator";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { LinkStatus, useLinkAccount } from "src/hooks/useLinkAccount";
import { App } from "src/types";

function AlbyConnectionCard({ connection }: { connection?: App }) {
  const { data: albyMe } = useAlbyMe();
  const { loading, linkStatus, loadingLinkStatus, linkAccount } =
    useLinkAccount();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="relative">
          Linked Alby Account
          {connection && <AppCardNotice app={connection} />}
        </CardTitle>
        <CardDescription>
          Link Your Alby Account to use your lightning address with Alby Hub and
          use apps that you connected to your Alby Account.
        </CardDescription>
      </CardHeader>
      <Separator />
      <CardContent className="group">
        <div className="grid grid-cols-1 xl:grid-cols-2 mt-5 gap-3 items-center relative">
          <div className="flex flex-col gap-4">
            <div className="flex flex-row gap-4">
              <UserAvatar className="h-14 w-14" />
              <div className="flex flex-col justify-center">
                <div className="text-xl font-semibold">
                  {albyMe?.name || albyMe?.email}
                </div>
                <div className="flex flex-row items-center gap-1 text-sm text-muted-foreground">
                  <ZapIcon className="w-4 h-4" />
                  {albyMe?.lightning_address}
                </div>
              </div>
            </div>
            <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
              {loadingLinkStatus && <Loading />}
              {!connection || linkStatus === LinkStatus.SharedNode ? (
                <LoadingButton onClick={linkAccount} loading={loading}>
                  {!loading && <Link2Icon className="w-4 h-4 mr-2" />}
                  Link your Alby Account
                </LoadingButton>
              ) : linkStatus === LinkStatus.ThisNode ? (
                <Button
                  variant="positive"
                  disabled
                  className="disabled:opacity-100"
                >
                  <CheckCircle2 className="w-4 h-4 mr-2" />
                  Alby Account Linked
                </Button>
              ) : (
                linkStatus === LinkStatus.OtherNode && (
                  <Button variant="destructive" disabled>
                    <CircleX className="w-4 h-4 mr-2" />
                    Linked to another wallet
                  </Button>
                )
              )}
              <ExternalLink
                to="https://www.getalby.com/node"
                className="w-full sm:w-auto"
              >
                <Button variant="outline" className="w-full sm:w-auto">
                  <ExternalLinkIcon className="w-4 h-4 mr-2" />
                  Alby Account Settings
                </Button>
              </ExternalLink>
            </div>
          </div>
          {connection && (
            <div>
              <Link
                to={`/apps/${connection.nostrPubkey}?edit=true`}
                className="absolute top-0 right-0"
              >
                <EditIcon className="w-4 h-4 hidden group-hover:inline text-muted-foreground hover:text-card-foreground" />
              </Link>
              <AppCardConnectionInfo connection={connection} />
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export default AlbyConnectionCard;
