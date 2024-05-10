import { ExternalLink } from "lucide-react";
import { NavLink, Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { buttonVariants } from "src/components/ui/button";
import { useInfo } from "src/hooks/useInfo";
import { cn } from "src/lib/utils";

export default function SettingsLayout() {
  const { data: info } = useInfo();
  return (
    <>
      <AppHeader
        title="Settings"
        description="Manage your Alby Hub settings"
      />
      <div className="flex flex-col space-y-8 lg:flex-row lg:space-x-12 lg:space-y-0">
        <aside className="-mx-4 lg:w-1/5">
          <nav className="flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1">
            <MenuItem to="/settings">General</MenuItem>
            <MenuItem to="/settings/change-unlock-password">
              Unlock Password
            </MenuItem>
            <MenuItem to="/settings/key-backup">Key Backup</MenuItem>
            {info?.backendType === "LDK" && (
              <MenuItem to="/settings/node-backup">Node Backup</MenuItem>
            )}
            <MenuItem to="/debug-tools">
              Debug Tools
              <ExternalLink className="w-4 h-4 ml-2" />
            </MenuItem>
          </nav>
        </aside>
        <div className="flex-1 lg:max-w-2xl">
          <div className="grid gap-6">
            <Outlet />
          </div>
        </div>
      </div>
    </>
  );
}

const MenuItem = ({
  to,
  children,
}: {
  to: string;
  children: React.ReactNode | string;
}) => (
  <>
    <NavLink
      end
      to={to}
      className={({ isActive }) =>
        cn(
          buttonVariants({ variant: "ghost" }),
          isActive
            ? "bg-muted hover:bg-muted"
            : "hover:bg-transparent hover:underline",
          "justify-start"
        )
      }
    >
      {children}
    </NavLink>
  </>
);

MenuItem;
