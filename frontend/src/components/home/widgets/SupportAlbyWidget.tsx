import { HeartIcon } from "lucide-react";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";
import { SUPPORT_ALBY_CONNECTION_NAME } from "src/constants";
import { useAlbyMe } from "src/hooks/useAlbyMe";
import { useApps } from "src/hooks/useApps";
import { useInfo } from "src/hooks/useInfo";

export function SupportAlbyWidget() {
  const { data: info } = useInfo();
  const { data: albyMe, error: albyMeError } = useAlbyMe();
  const { data: supportAlbyAppsData } = useApps(undefined, undefined, {
    name: SUPPORT_ALBY_CONNECTION_NAME,
  });

  if (
    !info ||
    (info.albyAccountConnected && !albyMe && !albyMeError) ||
    albyMe?.subscription.plan_code ||
    supportAlbyAppsData?.apps.length
  ) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Support Alby</CardTitle>
        <CardDescription>
          Upgrade to Pro or setup a recurring payment to support the development
          of Alby Hub, Alby Go and the NWC ecosystem.
        </CardDescription>
      </CardHeader>
      <CardFooter className="flex justify-end">
        <LinkButton to="/support-alby" variant="outline">
          <HeartIcon />
          Become a Supporter
        </LinkButton>
      </CardFooter>
    </Card>
  );
}
