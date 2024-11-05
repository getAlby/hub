import {
  CreditCard,
  Headphones,
  LifeBuoy,
  Mail,
  PartyPopper,
  Zap,
} from "lucide-react";
import Container from "src/components/Container";
import TwoColumnLayoutHeader from "src/components/TwoColumnLayoutHeader";
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
          description="Alby Account brings several benefits to your Alby Hub"
        />
        <div className="grid grid-cols-1 md:grid-cols-2 w-full gap-2 mt-5">
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <Zap className="w-6 h-6" />
              <CardTitle className="text-sm">Lightning Address</CardTitle>
              <CardDescription className="text-xs">
                Personalized lightning address to receive payments
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <Mail className="w-6 h-6" />
              <CardTitle className="text-sm">Email Notifications</CardTitle>
              <CardDescription className="text-xs">
                Instant notifications about incoming transactions and more
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <LifeBuoy className="w-6 h-6" />
              <CardTitle className="text-sm">Support</CardTitle>
              <CardDescription className="text-xs">
                Human support via live chat when you need a helping hand
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <CreditCard className="w-6 h-6" />
              <CardTitle className="text-sm">Fiat Topups</CardTitle>
              <CardDescription className="text-xs">
                Top up with fiat and get Bitcoin delivered to your Alby Hub
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <Headphones className="w-6 h-6" />
              <CardTitle className="text-sm">Podcasting 2.0</CardTitle>
              <CardDescription className="text-xs">
                Support your favorite creators by streaming sats
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <PartyPopper className="w-6 h-6" />
              <CardTitle className="text-sm">and there's more...</CardTitle>
              <CardDescription className="text-xs">
                Claim your Nostr address, discover apps, etc
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
        <div className="flex flex-col items-center justify-center mt-8 gap-2">
          <LinkButton to={connectUrl || "/alby/auth"} size="lg">
            Connect now
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
      </Container>
    </div>
  );
}
