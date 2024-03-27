import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
} from "src/components/ui/card";
import { Button } from "src/components/ui/button";
import { Link } from "react-router-dom";

function NotFound() {
  return (
    <div className="flex justify-center">
      <div className="container max-w-screen-lg">
        <Card>
          <CardHeader>
            <CardTitle>Page Not Found</CardTitle>
            <CardDescription>
              The page you are looking for does not exist.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Link to="/">
              <Button variant="link">Return Home</Button>
            </Link>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default NotFound;
