import { ExternalLinkIcon } from "lucide-react";
import { AlbyHead } from "src/components/images/AlbyHead";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { ExternalLinkButton } from "src/components/ui/custom/external-link-button";
import { useInfo } from "src/hooks/useInfo";

export function AlbyAccountWidget() {
  const { data: info } = useInfo();

  if (!info || !info.albyAccountConnected) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-row items-center">
          <div className="shrink-0">
            <AlbyHead className="h-12 w-12 rounded-xl p-1 border" />
          </div>
          <div>
            <CardTitle>
              <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                Alby Account
              </div>
            </CardTitle>
            <CardDescription className="ml-4">
              Get an Alby Account with a web wallet interface, lightning address
              and other features.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardFooter className="justify-end px-6 pt-0">
        <ExternalLinkButton
          to="https://www.getalby.com/dashboard"
          variant="outline"
        >
          Open Alby Account
          <ExternalLinkIcon className="size-4" />
        </ExternalLinkButton>
      </CardFooter>
    </Card>
  );
}
