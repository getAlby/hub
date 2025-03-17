import { ReactElement } from "react";
import Breadcrumbs from "src/components/Breadcrumbs";
import { SidebarTrigger } from "src/components/ui/sidebar";

type Props = {
  title: string | ReactElement;
  description?: string | ReactElement;
  contentRight?: React.ReactNode;
  breadcrumb?: boolean;
};

function AppHeader({
  title,
  description = "",
  contentRight,
  breadcrumb = true,
}: Props) {
  return (
    <>
      <div className="flex flex-row items-center border-b border-border pb-3 lg:pb-6 gap-5">
        <SidebarTrigger />
        <div className="flex flex-col flex-1">
          {breadcrumb && <Breadcrumbs />}
          <div className="flex justify-between items-center">
            <div className="flex-1">
              <h1 className="text-xl lg:text-3xl font-semibold">{title}</h1>
              <p className="hidden lg:inline text-muted-foreground">
                {description}
              </p>
            </div>
            <div className="flex gap-3 h-full">{contentRight}</div>
          </div>
        </div>
      </div>
    </>
  );
}

export default AppHeader;
