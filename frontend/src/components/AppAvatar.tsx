import { suggestedApps } from "src/components/connections/SuggestedAppData";
import UserAvatar from "src/components/UserAvatar";
import { ALBY_ACCOUNT_APP_NAME } from "src/constants";
import { cn } from "src/lib/utils";
import { App } from "src/types";

type Props = {
  app: App;
  className?: string;
};

export default function AppAvatar({ app, className }: Props) {
  if (app.name === ALBY_ACCOUNT_APP_NAME) {
    return <UserAvatar className={className} />;
  }
  const appStoreApp = suggestedApps.find(
    (suggestedApp) =>
      (app?.metadata?.app_store_app_id &&
        suggestedApp.id === app.metadata?.app_store_app_id) ||
      app.name.includes(suggestedApp.title)
  );
  const image = appStoreApp?.logo;

  const gradient =
    app.name
      .split("")
      .map((c) => c.charCodeAt(0))
      .reduce((a, b) => a + b, 0) % 10;
  return (
    <div
      className={cn(
        "rounded-lg border relative overflow-hidden",
        !image && `avatar-gradient-${gradient}`,
        className
      )}
    >
      {image && (
        <img
          src={image}
          className={cn("absolute w-full h-full rounded-lg", className)}
        />
      )}
      {!image && (
        <span className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-white text-sm font-medium capitalize pointer-events-none">
          {app.name.charAt(0)}
        </span>
      )}
    </div>
  );
}
