import gradientAvatar from "gradient-avatar";
import { cn } from "src/lib/utils";

type Props = {
  appName: string;
  className?: string;
};

export default function AppAvatar({ appName, className }: Props) {
  return (
    <div className={cn("rounded-lg border relative", className)}>
      <img
        src={`data:image/svg+xml;base64,${btoa(gradientAvatar(appName))}`}
        alt={appName}
        className="block w-full h-full rounded-lg p-1"
      />
      <span className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-white text-xl font-medium capitalize">
        {appName.charAt(0)}
      </span>
    </div>
  );
}
