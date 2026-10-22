import React from "react";
import { Link, useSearchParams } from "react-router";
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
  getAppStoreUrl,
  sortedAppStoreCategories,
} from "./SuggestedAppData";

function AppCard(app: AppStoreApp) {
  return (
    <Link to={getAppStoreUrl(app)} className="min-w-0">
      <Card className="h-full min-w-0">
        <CardContent>
          <div className="flex gap-3 items-center">
            <img
              src={app.logo}
              alt={`${app.title} logo`}
              className="inline rounded-lg size-12"
            />
            <div className="grow min-w-0">
              <CardTitle>{app.title}</CardTitle>
              <CardDescription>
                {app.description}
                {app.legacyTitles?.length
                  ? ` (Previously ${app.legacyTitles.join(", ")})`
                  : ""}
              </CardDescription>
            </div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

export default function SuggestedApps() {
  const [searchParams] = useSearchParams();
  const [selectedCategories, setSelectedCategories] = React.useState<string[]>(
    () => {
      const category = searchParams.get("category");
      return category ? [category] : [];
    }
  );

  return (
    <>
      <div className="flex gap-2 flex-wrap">
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
      <div className="flex flex-col gap-4">
        {sortedAppStoreCategories
          .filter(
            ([categoryId]) =>
              !selectedCategories.length ||
              selectedCategories.includes(categoryId)
          )
          .map(([categoryId, category]) => {
            return (
              <div key={categoryId}>
                <h3 className="font-semibold text-xl">{category.title}</h3>
                <div className="grid md:grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4 mt-4">
                  {appStoreApps
                    .filter((app) =>
                      (app.categories as string[]).includes(categoryId)
                    )
                    .map((app) => (
                      <AppCard key={app.id} {...app} />
                    ))}
                </div>
              </div>
            );
          })}
      </div>
    </>
  );
}
