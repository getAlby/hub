import { SearchXIcon } from "lucide-react";
import { Link } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "src/components/ui/card";

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
        <Link to="/">
          <Button variant="outline">Return Home</Button>
        </Link>
      </CardContent>
    </Card>
  );
}

export default NotFound;
