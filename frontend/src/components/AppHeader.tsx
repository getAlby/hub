import { ReactElement } from "react";
import { Separator } from "src/components/ui/separator";
import { SidebarTrigger } from "src/components/ui/sidebar";

type Props = {
  title: string | ReactElement;
  description?: string | ReactElement;
  contentRight?: React.ReactNode;
  breadcrumb?: boolean;
};

function AppHeader({ title, description = "", contentRight }: Props) {
  return (
    <>
      <header className="flex flex-row items-center border-b border-border pb-3 lg:pb-6 gap-2">
        <SidebarTrigger className="-ml-1" />
        <Separator orientation="vertical" className="mr-2 h-4" />
        <div className="flex flex-col flex-1">
          {/* {breadcrumb && <Breadcrumbs />} */}
          <div className="flex justify-between items-center">
            <div className="flex-1">
              <h1 className="text-2xl lg:text-3xl font-semibold">{title}</h1>
              {description && (
                <p className="hidden lg:inline text-muted-foreground">
                  {description}
                </p>
              )}
            </div>
            <div className="flex gap-3 h-full">{contentRight}</div>
          </div>
        </div>
      </header>
    </>
  );
}

export default AppHeader;
