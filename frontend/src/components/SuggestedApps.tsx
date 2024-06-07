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
    <Link to={`/apps/new?app=${id}`}>
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-3 items-center">
            <img
              src={logo}
              alt="logo"
              className="inline rounded-lg w-10 h-10"
            />
            <div className="flex-grow">
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
      <div className="grid sm:grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
        {suggestedApps.map((app) => (
          <SuggestedAppCard
            id={app.id}
            key={app.id}
            to={app.to}
            title={app.title}
            description={app.description}
            logo={app.logo}
          />
        ))}
      </div>
    </>
  );
}
