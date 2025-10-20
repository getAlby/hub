import React from "react";
import { Link } from "react-router-dom";
import { Badge } from "src/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardTitle,
} from "src/components/ui/card";
import { cn } from "src/lib/utils";
import {
  AppStoreApp,
  appStoreApps,
  sortedAppStoreCategories,
} from "./SuggestedAppData";

function SuggestedAppCard({ id, title, description, logo }: AppStoreApp) {
  return (
    <Link to={`/appstore/${id}`}>
      <Card className="h-full">
        <CardContent>
          <div className="flex gap-3 items-center">
            <img src={logo} alt="logo" className="inline rounded-lg size-12" />
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

function InternalAppCard({ id, title, description, logo }: AppStoreApp) {
  return (
    <Link to={`/internal-apps/${id}`}>
      <Card className="h-full">
        <CardContent>
          <div className="flex gap-3 items-center">
            <img src={logo} alt="logo" className="inline rounded-lg size-12" />
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
  const [selectedCategories, setSelectedCategories] = React.useState<string[]>(
    []
  );

  return (
    <>
      <div className="flex gap-2 flex-wrap mt-6 mb-2">
        {sortedAppStoreCategories.map(([categoryId, category]) => (
          <Badge
            key={categoryId}
            variant={
              selectedCategories.includes(categoryId) ? "default" : "secondary"
            }
            className={cn(
              "cursor-pointer",
              selectedCategories.includes(categoryId)
                ? ""
                : "border-transparent font-normal select-none"
            )}
            onClick={() =>
              setSelectedCategories((current) => [
                ...current.filter((c) => c !== categoryId),
                ...(current.includes(categoryId) ? [] : [categoryId]),
              ])
            }
          >
            {category.title}
          </Badge>
        ))}
      </div>
      <div className="flex flex-col gap-8">
        {sortedAppStoreCategories
          .filter(
            ([categoryId]) =>
              !selectedCategories.length ||
              selectedCategories.includes(categoryId)
          )
          .map(([categoryId, category]) => {
            return (
              <div key={categoryId} className="pt-4">
                <h3 className="font-semibold text-xl">{category.title}</h3>
                <div className="grid md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4 mt-4">
                  {appStoreApps
                    .filter((app) =>
                      (app.categories as string[]).includes(categoryId)
                    )
                    .map((app) =>
                      app.internal ? (
                        <InternalAppCard key={app.id} {...app} />
                      ) : (
                        <SuggestedAppCard key={app.id} {...app} />
                      )
                    )}
                </div>
              </div>
            );
          })}
      </div>
    </>
  );
}
