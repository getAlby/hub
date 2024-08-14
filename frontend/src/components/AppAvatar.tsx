import { cn } from "src/lib/utils";

type Props = {
  appName: string;
  className?: string;
};

export default function AppAvatar({ appName, className }: Props) {
  const gradient =
    appName
      .split("")
      .map((c) => c.charCodeAt(0))
      .reduce((a, b) => a + b, 0) % 10;
  return (
    <div
      className={cn(
        "rounded-lg border relative",
        `avatar-gradient-${gradient}`,
        className
      )}
    >
      <span className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 text-white text-xl font-medium capitalize">
        {appName.charAt(0)}
      </span>
    </div>
  );
}
