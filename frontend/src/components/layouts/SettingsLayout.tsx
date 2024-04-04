import { NavLink, Outlet } from "react-router-dom";
import AppHeader from "src/components/AppHeader";
import { buttonVariants } from "src/components/ui/button";
import { cn } from "src/lib/utils";

export default function SettingsLayout() {
  return (
    <>
      <AppHeader
        title="Settings"
        description="Manage your account settings and set e-mail preferences."
      />
      <div className="grid w-full items-start gap-6 md:grid-cols-[180px_1fr] lg:grid-cols-[250px_1fr]">
        <nav className="flex space-x-2 lg:flex-col lg:space-x-0 lg:space-y-1">
          <MenuItem to="/settings">General</MenuItem>
          <MenuItem to="/settings/backup">Backup</MenuItem>
          <MenuItem to="/settings/change-unlock-password">
            Unlock Password
          </MenuItem>
        </nav>
        <div className="grid gap-6">
          <Outlet />
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
