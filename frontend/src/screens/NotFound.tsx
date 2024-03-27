import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "src/components/ui/card";
import { Button } from "src/components/ui/button";
import { Link } from "react-router-dom";
import { SearchX } from "lucide-react";

function NotFound() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>
          <div className="flex flex-row items-center gap-2">
            <SearchX className="w-10 h-10" />
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
