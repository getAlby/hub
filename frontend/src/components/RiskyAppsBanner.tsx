import { XIcon } from "lucide-react";
import React from "react";
import { Link } from "react-router-dom";
import { useRiskyApps } from "src/hooks/useRiskyApps";

export function RiskyAppsBanner() {
  const data = useRiskyApps();
  const [dismissed, setDismissed] = React.useState(false);

  if (!data?.apps?.length || dismissed) {
    return null;
  }

  const count = data.apps.length;
  const appNames = data.apps
    .map((app: { name: string }) => app.name)
    .slice(0, 3)
    .join(", ");
  const hasMore = data.apps.length > 3;

  return (
    <div className="fixed w-full bg-orange-500 dark:bg-orange-600 text-white z-20 py-2 text-sm flex items-center justify-center">
      <Link to="/apps" className="w-full px-12 md:px-24">
        <p className="line-clamp-2 md:block whitespace-normal md:whitespace-nowrap overflow-hidden text-ellipsis text-center">
          <span className="font-semibold mr-2">
            {count} app{count > 1 ? "s" : ""} with spending permissions{" "}
            {count === 1 ? "hasn't" : "haven't"}
            been used in 30+ days
          </span>
          <span>•</span>
          <span className="ml-2">
            {appNames}
            {hasMore ? `, and ${data.apps.length - 3} more` : ""}
          </span>
        </p>
      </Link>
      <XIcon
        className="absolute right-4 cursor-pointer w-4 text-white"
        role="button"
        aria-label="Dismiss banner"
        tabIndex={0}
        onClick={(e) => {
          e.preventDefault();
          setDismissed(true);
        }}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            setDismissed(true);
          }
        }}
      />
    </div>
  );
}
