import { ExternalLinkIcon } from "lucide-react";
import AppHeader from "src/components/AppHeader";
import ExternalLink from "src/components/ExternalLink";
import { AlbyHead } from "src/components/images/AlbyHead";
import Loading from "src/components/Loading";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useBalances } from "src/hooks/useBalances";
import { useInfo } from "src/hooks/useInfo";
import OnboardingChecklist from "src/screens/wallet/OnboardingChecklist";

function getGreeting(name: string | undefined) {
  const hours = new Date().getHours();
  let greeting;

  if (hours < 11) {
    greeting = "Good Morning";
  } else if (hours < 16) {
    greeting = "Good Afternoon";
  } else {
    greeting = "Good Evening";
  }

  return `${greeting}${name && `, ${name}`}!`;
}

function Home() {
  const { data: info } = useInfo();
  const { data: balances } = useBalances();
  const { data: albyMe } = useAlbyMe();

  /* eslint-disable  @typescript-eslint/no-explicit-any */
  const extensionInstalled = (window as any).alby !== undefined;

  if (!info || !balances || !albyMe) {
    return <Loading />;
  }

  return (
    <>
      <AppHeader title={getGreeting(albyMe?.name)} description="" />
      <OnboardingChecklist />
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5">
        <ExternalLink to="https://www.getalby.com/dashboard">
          <Card>
            <CardHeader>
              <div className="flex flex-row items-center">
                <div className="flex-shrink-0">
                  <AlbyHead className="w-12 h-12 rounded-xl p-1 border" />
                </div>
                <div>
                  <CardTitle>
                    <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                      Alby Web
                    </div>
                  </CardTitle>
                  <CardDescription className="ml-4">
                    Install Alby Web on your phone and use your Hub on the go.
                  </CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="text-right">
              <Button variant="outline">
                Open Alby Web
                <ExternalLinkIcon className="w-4 h-4 ml-2" />
              </Button>
            </CardContent>
          </Card>
        </ExternalLink>
        {!extensionInstalled && (
          <ExternalLink to="https://www.getalby.com">
            <Card>
              <CardHeader>
                <div className="flex flex-row items-center">
                  <div className="flex-1">
                    <AlbyHead className="w-12 h-12 rounded-xl p-1 border bg-[#FFDF6F]" />
                  </div>
                  <div>
                    <CardTitle>
                      <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                        Alby Browser Extension
                      </div>
                    </CardTitle>
                    <CardDescription className="ml-4">
                      Seamless bitcoin payments in your favourite internet
                      browser.
                    </CardDescription>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="text-right">
                <Button variant="outline">
                  Install Alby Extension
                  <ExternalLinkIcon className="w-4 h-4 ml-2" />
                </Button>
              </CardContent>
            </Card>
          </ExternalLink>
        )}
      </div>
    </>
  );
}

export default Home;
