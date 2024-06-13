import { Avatar, AvatarFallback, AvatarImage } from "@radix-ui/react-avatar";
import { Separator } from "@radix-ui/react-dropdown-menu";
import { Progress } from "@radix-ui/react-progress";
import {
  CheckCircle2,
  CircleX,
  Edit,
  ExternalLinkIcon,
  Link2Icon,
  ZapIcon,
} from "lucide-react";
import { Link } from "react-router-dom";
import ExternalLink from "src/components/ExternalLink";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LoadingButton } from "src/components/ui/loading-button";
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
        <CardTitle>Alby Account</CardTitle>
        <CardDescription>
          Link Your Alby Account to use your lightning address with Alby Hub and
          use apps that you connected to your Alby Account.
        </CardDescription>
      </CardHeader>
      <Separator />
      <CardContent>
        <div className="grid grid-cols-1 xl:grid-cols-2 mt-5 gap-3 items-center">
          <div className="flex flex-col gap-4">
            <div className="flex flex-row gap-4 ">
              <Avatar className="h-14 w-14">
                <AvatarImage src={albyMe?.avatar} alt="@satoshi" />
                <AvatarFallback>SN</AvatarFallback>
              </Avatar>
              <div className="flex flex-col">
                <div className="text-xl font-semibold">{albyMe?.name}</div>
                <div className="flex flex-row items-center gap-1 text-sm">
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
                className="w-full md:w-auto"
              >
                <Button variant="outline" className="w-full md:w-auto">
                  <ExternalLinkIcon className="w-4 h-4 mr-2" />
                  Alby Account Settings
                </Button>
              </ExternalLink>
            </div>
          </div>
          <div>
            {connection && (
              <>
                {connection.maxAmount > 0 && (
                  <>
                    <div className="flex flex-row justify-between">
                      <div className="mb-2">
                        <p className="text-xs text-secondary-foreground font-medium">
                          You've spent
                        </p>
                        <p className="text-xl font-medium">
                          {new Intl.NumberFormat().format(
                            connection.budgetUsage
                          )}{" "}
                          sats
                        </p>
                      </div>
                      <div className="text-right">
                        {" "}
                        <p className="text-xs text-secondary-foreground font-medium">
                          Left in budget
                        </p>
                        <p className="text-xl font-medium text-muted-foreground">
                          {new Intl.NumberFormat().format(
                            connection.maxAmount - connection.budgetUsage
                          )}{" "}
                          sats
                        </p>
                      </div>
                    </div>
                    <Progress
                      className="h-4"
                      value={
                        (connection.budgetUsage * 100) / connection.maxAmount
                      }
                    />
                    <div className="flex flex-row justify-between text-xs items-center mt-2">
                      {connection.maxAmount > 0 ? (
                        <>
                          {new Intl.NumberFormat().format(connection.maxAmount)}{" "}
                          sats / {connection.budgetRenewal}
                        </>
                      ) : (
                        "Not set"
                      )}
                      <div>
                        <Link to={`/apps/${connection.nostrPubkey}`}>
                          <Button variant="ghost" size="sm">
                            <Edit className="w-4 h-4 mr-2" />
                            Edit
                          </Button>
                        </Link>
                      </div>
                    </div>
                  </>
                )}
              </>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default AlbyConnectionCard;
