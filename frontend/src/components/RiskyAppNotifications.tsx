import { BellIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Button } from "src/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "src/components/ui/popover";
import { useRiskyApps } from "src/hooks/useRiskyApps";

export function RiskyAppNotifications() {
  const data = useRiskyApps();
  const navigate = useNavigate();
  const [minimized, setMinimized] = useState(false);
  const [popoverOpen, setPopoverOpen] = useState(false);

  const hasRiskyApps = data?.apps?.length && data.apps.length > 0;
  const count = data?.apps?.length || 0;

  useEffect(() => {
    if (hasRiskyApps) {
      // Auto-minimize after 5 seconds
      const timer = setTimeout(() => {
        setMinimized(true);
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [hasRiskyApps]);

  const handleNotificationClick = () => {
    console.info("🔔 Notification clicked! Navigating to /apps");
    console.info("Has risky apps:", hasRiskyApps);
    console.info("Count:", count);
    setPopoverOpen(false);
    navigate("/apps");
  };

  return (
    <>
      {/* Toast Notification */}
      {hasRiskyApps && (
        <div
          className={`fixed top-4 right-4 z-50 transition-all duration-500 transform ${
            !minimized
              ? "translate-y-0 opacity-100"
              : "-translate-y-4 opacity-0 pointer-events-none"
          }`}
        >
          <div className="bg-popover text-popover-foreground border shadow-lg rounded-lg p-4 w-80 flex flex-col gap-2">
            <div className="font-semibold flex justify-between items-center">
              <span>Security Alert</span>
              <button
                type="button"
                aria-label="Dismiss notification"
                onClick={() => setMinimized(true)}
                className="text-muted-foreground hover:text-foreground"
              >
                ×
              </button>
            </div>
            <p className="text-sm">
              {count} app{count > 1 ? "s" : ""} {count === 1 ? "has" : "have"}{" "}
              unused spending permissions.
            </p>
            <div className="flex justify-end gap-2 mt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setMinimized(true)}
              >
                Dismiss
              </Button>
              <Button size="sm" asChild>
                <Link to="/apps">Review</Link>
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Notification Icon (Bell) */}
      <Popover modal={true} open={popoverOpen} onOpenChange={setPopoverOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="relative"
            aria-label="Notifications"
            onClick={() => {
              console.info("🔔 Bell icon clicked!");
              console.info("Current popover state:", popoverOpen);
            }}
          >
            <BellIcon className="h-5 w-5" />
            {/* Badge */}
            <span
              className={`absolute top-2 right-2 h-2 w-2 rounded-full bg-orange-500 transition-transform duration-300 ${
                hasRiskyApps && minimized ? "scale-100" : "scale-0"
              }`}
            />
          </Button>
        </PopoverTrigger>
        <PopoverContent
          className="w-80"
          align="end"
          side="bottom"
          sideOffset={8}
        >
          <div className="grid gap-4">
            <div className="space-y-2">
              <h4 className="font-medium leading-none">Notifications</h4>
              <p className="text-sm text-muted-foreground">
                Security alerts and updates.
              </p>
            </div>
            <div className="grid gap-2">
              {hasRiskyApps ? (
                <button
                  type="button"
                  onClick={handleNotificationClick}
                  className="flex items-start gap-4 rounded-md border p-3 hover:bg-muted/50 cursor-pointer transition-colors w-full text-left"
                  aria-label="Review apps with unused permissions"
                >
                  <div className="flex-1 space-y-1">
                    <p className="text-sm font-medium leading-none text-orange-600">
                      Unused Permissions
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {count} app{count > 1 ? "s" : ""} can spend but{" "}
                      {count === 1 ? "hasn't" : "haven't"}
                      been used recently.
                    </p>
                    <span className="text-xs font-medium underline text-primary">
                      Review Apps
                    </span>
                  </div>
                </button>
              ) : (
                <div className="py-6 text-center text-sm text-muted-foreground">
                  No new notifications
                </div>
              )}
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </>
  );
}
