import albyGo from "src/assets/suggested-apps/alby-go.png";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";

export function AlbyGoWidget() {
  return (
    <Card>
      <CardHeader>
        <div className="flex flex-row items-center">
          <div className="shrink-0">
            <img
              src={albyGo}
              alt="Alby Go"
              className="h-12 w-12 rounded-xl border"
            />
          </div>
          <div>
            <CardTitle>
              <div className="flex-1 leading-5 font-semibold text-xl whitespace-nowrap text-ellipsis overflow-hidden ml-4">
                Alby Go
              </div>
            </CardTitle>
            <CardDescription className="ml-4">
              The easiest Bitcoin mobile app that works great with Alby Hub.
            </CardDescription>
          </div>
        </div>
      </CardHeader>
      <CardContent className="text-right">
        <LinkButton to="/appstore/alby-go" variant="outline">
          Open
        </LinkButton>
      </CardContent>
    </Card>
  );
}
