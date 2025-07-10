import {
  CreditCardIcon,
  DatabaseBackupIcon,
  HeadphonesIcon,
  LifeBuoyIcon,
  MailIcon,
  SparklesIcon,
  ZapIcon,
} from "lucide-react";
import Container from "src/components/Container";
import ExternalLink from "src/components/ExternalLink";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
import { Badge } from "src/components/ui/badge";
import { LinkButton } from "src/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

type ConnectAlbyAccountProps = {
  connectUrl?: string;
};

export function ConnectAlbyAccount({ connectUrl }: ConnectAlbyAccountProps) {
  return (
    <div className="w-full h-full flex flex-col items-center justify-center gap-5">
      <Container>
        <TwoColumnLayoutHeader
          title="Connect Your Alby Account"
          description="Your Alby Account brings several benefits to your Alby Hub"
        />
        <div className="grid grid-cols-1 md:grid-cols-2 w-full gap-2 mt-5">
          <Card className="w-full relative">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <ZapIcon className="size-6" />
              <CardTitle className="text-sm">
                Lightning Address
                <AlbyProIcon />
              </CardTitle>
              <CardDescription className="text-xs">
                Personalized lightning address to receive payments
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full relative">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <MailIcon className="size-6" />
              <CardTitle className="text-sm">
                Email Notifications
                <AlbyProIcon />
              </CardTitle>
              <CardDescription className="text-xs">
                Instant notifications about incoming transactions and more
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full relative">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <DatabaseBackupIcon className="size-6" />
              <CardTitle className="text-sm">
                Encrypted Backups
                <AlbyProIcon />
              </CardTitle>
              <CardDescription className="text-xs">
                Ensures you can always recover funds from lightning channels
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full relative">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <LifeBuoyIcon className="size-6" />
              <CardTitle className="text-sm">
                Support
                <AlbyProIcon />
              </CardTitle>
              <CardDescription className="text-xs">
                Human support via live chat when you need a helping hand
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <CreditCardIcon className="size-6" />
              <CardTitle className="text-sm">Fiat Topups</CardTitle>
              <CardDescription className="text-xs">
                Top up with fiat and get Bitcoin delivered to your Alby Hub
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <HeadphonesIcon className="size-6" />
              <CardTitle className="text-sm">Podcasting 2.0</CardTitle>
              <CardDescription className="text-xs">
                Support your favorite creators by streaming sats
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
        <div className="flex flex-col items-center justify-center gap-2 mt-10">
          <LinkButton to={connectUrl || "/alby/auth"} size="lg">
            Connect
          </LinkButton>
          <LinkButton
            size="sm"
            variant="link"
            to="/"
            className="text-muted-foreground"
          >
            Maybe later
          </LinkButton>
        </div>
        <div className="text-muted-foreground flex flex-col items-center text-xs gap-2 mt-10">
          <Badge title="Pro" variant="outline">
            <SparklesIcon className="size-4" />
          </Badge>
          <div>
            Unlock additional features with{" "}
            <ExternalLink
              to="https://getalby.com/pricing"
              className="underline"
            >
              Pro
            </ExternalLink>
          </div>
        </div>
      </Container>
    </div>
  );

  function AlbyProIcon() {
    return (
      <div className="absolute right-2 top-2" title="Pro">
        <SparklesIcon className="size-4" />
      </div>
    );
  }
}
