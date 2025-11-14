import { SearchXIcon } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";
import { LinkButton } from "src/components/ui/custom/link-button";

function NotFound() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          <div className="flex flex-row items-center gap-2">
            <SearchXIcon className="w-10 h-10" />
            Page Not Found
          </div>
        </CardTitle>
        <CardDescription>
          The page you are looking for does not exist.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <LinkButton to="/" variant="outline">
          Return Home
        </LinkButton>
      </CardContent>
    </Card>
  );
}

export default NotFound;
