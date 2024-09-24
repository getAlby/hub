import {
  CreditCard,
  DatabaseBackup,
  Headphones,
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

export function ConnectAlbyAccount() {
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
                Personalized lightning address to receive payments straight into
                your Hub
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <Mail className="w-6 h-6" />
              <CardTitle className="text-sm">Email Notifications</CardTitle>
              <CardDescription className="text-xs">
                Stay updated! Get instant emails about incoming transactions and
                important updates.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <DatabaseBackup className="w-6 h-6" />
              <CardTitle className="text-sm">Encrypted Backups</CardTitle>
              <CardDescription className="text-xs">
                Enjoy peace of mind with automated backups of your <br />
                Alby Hub.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <CreditCard className="w-6 h-6" />
              <CardTitle className="text-sm">Fiat Topups</CardTitle>
              <CardDescription className="text-xs">
                Top up with fiat and get Bitcoin delivered straight into your{" "}
                <br />
                Alby Hub.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <Headphones className="w-6 h-6" />
              <CardTitle className="text-sm">Podcasting 2.0</CardTitle>
              <CardDescription className="text-xs">
                Support your favorite creators by streaming sats while you enjoy
                their content.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card className="w-full">
            <CardHeader className="flex flex-col justify-center items-center text-center p-4">
              <PartyPopper className="w-6 h-6" />
              <CardTitle className="text-sm">and there's more...</CardTitle>
              <CardDescription className="text-xs">
                Claim your Nostr address, discover new apps and explore what's
                coming next.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
        <div className="flex flex-col items-center justify-center mt-8 gap-2">
          <LinkButton to="/alby/auth" size="lg">
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
