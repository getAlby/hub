import { Link } from "react-router-dom";
import {
  Card,
  CardContent,
  CardDescription,
  CardTitle,
} from "src/components/ui/card";
import { SuggestedApp, suggestedApps } from "./SuggestedAppData";

function SuggestedAppCard({ id, title, description, logo }: SuggestedApp) {
  return (
    <Link to={`/appstore/${id}`}>
      <Card className="h-full">
        <CardContent className="p-4">
          <div className="flex gap-3 items-center">
            <img
              src={logo}
              alt="logo"
              className="inline rounded-lg w-12 h-12"
            />
            <div className="grow">
              <CardTitle>{title}</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

function InternalAppCard({ id, title, description, logo }: SuggestedApp) {
  return (
    <Link to={`/internal-apps/${id}`}>
      <Card className="h-full">
        <CardContent className="p-4">
          <div className="flex gap-3 items-center">
            <img
              src={logo}
              alt="logo"
              className="inline rounded-lg w-12 h-12"
            />
            <div className="grow">
              <CardTitle>{title}</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

export default function SuggestedApps() {
  return (
    <>
      <div className="grid md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
        {suggestedApps.map((app) =>
          app.internal ? (
            <InternalAppCard key={app.id} {...app} />
          ) : (
            <SuggestedAppCard key={app.id} {...app} />
          )
        )}
      </div>
    </>
  );
}
