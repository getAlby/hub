import { ReactElement } from "react";
import { Separator } from "src/components/ui/separator";
import { SidebarTrigger } from "src/components/ui/sidebar";

type Props = {
  title: string | ReactElement;
  description?: string | ReactElement;
  contentRight?: React.ReactNode;
  contentRightOnNewLine?: boolean;
  breadcrumb?: boolean;
  addSidebarTrigger?: boolean;
};

function AppHeader({
  title,
  description = "",
  contentRight,
  addSidebarTrigger = true,
}: Props) {
  return (
    <>
      <header className="flex flex-row flex-wrap items-center border-b border-border pb-4 gap-2">
        {addSidebarTrigger && <SidebarTrigger className="-ml-1 md:hidden" />}
        <Separator orientation="vertical" className="mr-2 h-4 md:hidden" />
        <div className="flex flex-col flex-1">
          <div className="flex justify-between items-center flex-wrap gap-2">
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
