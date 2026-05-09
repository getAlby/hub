import { HeartIcon } from "lucide-react";
import {
  Card,
  CardContent,
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
      <CardHeader className="px-6 pb-0">
        <CardTitle className="text-base font-semibold">
          Support Open Source
        </CardTitle>
      </CardHeader>
      <CardContent className="px-6 py-0">
        <div className="flex items-center gap-4">
          <div className="flex size-12 items-center justify-center text-orange-500">
            <HeartIcon className="size-12 stroke-[1.75]" />
          </div>
          <CardDescription className="text-sm leading-5">
            Upgrade to Pro or donate to support the development of Alby Hub,
            Alby Go and Alby Extension.
          </CardDescription>
        </div>
      </CardContent>
      <CardFooter className="justify-end px-6 pt-0">
        <LinkButton to="/support-alby" variant="outline">
          Become a Supporter
        </LinkButton>
      </CardFooter>
    </Card>
  );
}
