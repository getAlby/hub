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

export function AlbyExtensionWidget() {
  /* eslint-disable  @typescript-eslint/no-explicit-any */
  const extensionInstalled = (window as any).alby !== undefined;

  if (extensionInstalled) {
    return null;
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-row items-center">
          <div className="shrink-0">
            <AlbyHead className="w-12 h-12 rounded-xl p-1 border bg-[#FFDF6F]" />
          </div>
          <div>
            <CardTitle>
              <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                Alby Browser Extension
              </div>
            </CardTitle>
            <CardDescription className="ml-4">
              Seamless bitcoin payments in your favorite internet browser.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardFooter className="justify-end px-6 pt-0">
        <ExternalLinkButton
          to="https://getalby.com/products/browser-extension"
          variant="outline"
        >
          Install Alby Extension
          <ExternalLinkIcon className="size-4" />
        </ExternalLinkButton>
      </CardFooter>
    </Card>
  );
}
