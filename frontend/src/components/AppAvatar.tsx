import claudeLogo from "src/assets/suggested-apps/claude.png";
import clineLogo from "src/assets/suggested-apps/cline.png";
import codexLogo from "src/assets/suggested-apps/codex.png";
import cursorLogo from "src/assets/suggested-apps/cursor.png";
import geminiLogo from "src/assets/suggested-apps/gemini.png";
import gooseLogo from "src/assets/suggested-apps/goose.png";
import openclawLogo from "src/assets/suggested-apps/openclaw.png";
import opencodeLogo from "src/assets/suggested-apps/opencode.png";
import { appStoreApps } from "src/components/connections/SuggestedAppData";
import UserAvatar from "src/components/UserAvatar";
import { ALBY_ACCOUNT_APP_NAME } from "src/constants";
import { cn } from "src/lib/utils";
import { App } from "src/types";

// Lightweight logo map for agents that don't have full app store entries.
// Matches on app_store_app_id metadata set when the connection was created.
const agentLogos: Record<string, string> = {
  claude: claudeLogo,
  goose: gooseLogo,
  openclaw: openclawLogo,
  cursor: cursorLogo,
  codex: codexLogo,
  cline: clineLogo,
  gemini: geminiLogo,
  opencode: opencodeLogo,
};

type Props = {
  app: App;
  className?: string;
};

export default function AppAvatar({ app, className }: Props) {
  if (app.name === ALBY_ACCOUNT_APP_NAME) {
    return <UserAvatar className={className} />;
  }
  const agentLogo =
    (app?.metadata?.app_store_app_id
      ? agentLogos[app.metadata.app_store_app_id]
      : undefined) ||
    Object.entries(agentLogos).find(([key]) =>
      app.name.toLowerCase().includes(key)
    )?.[1];
  const appStoreApp = appStoreApps.find(
    (suggestedApp) =>
      (app?.metadata?.app_store_app_id &&
        suggestedApp.id === app.metadata?.app_store_app_id) ||
      app.name.includes(suggestedApp.title)
  );
  const image = agentLogo || appStoreApp?.logo;

  const gradient =
    app.name
      .split("")
      .map((c) => c.charCodeAt(0))
      .reduce((a, b) => a + b, 0) % 10;
  return (
    <div
      className={cn(
        "rounded-lg relative overflow-hidden",
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
