import { ReactElement } from "react";
import { Separator } from "src/components/ui/separator";
import { SidebarTrigger } from "src/components/ui/sidebar";

type Props = {
  icon?: React.ReactElement;
  title: string | ReactElement;
  description?: string | ReactElement;
  contentRight?: React.ReactNode;
  breadcrumb?: boolean;
  addSidebarTrigger?: boolean;
  pageTitle?: string;
};

function AppHeader({
  icon,
  title,
  description = "",
  contentRight,
  addSidebarTrigger = true,
  pageTitle,
}: Props) {
  return (
    <>
      {pageTitle && <title>{`${pageTitle} - Alby Hub`}</title>}
      <header className="flex flex-row flex-wrap items-center border-b border-border pb-4 gap-2">
        {addSidebarTrigger && <SidebarTrigger className="-ml-1 md:hidden" />}
        <Separator orientation="vertical" className="mr-2 h-4 md:hidden" />
        {icon}
        <div className="flex flex-col flex-1 min-w-0">
          <div className="flex justify-between items-start sm:items-center flex-wrap gap-2">
            <div className="flex-1 min-w-0">
              <h1 className="text-2xl lg:text-3xl font-semibold">{title}</h1>
              {description && (
                <p className="text-xs sm:text-base text-muted-foreground">
                  {description}
                </p>
              )}
            </div>
            {contentRight && (
              <div className="flex gap-3 h-full shrink-0">{contentRight}</div>
            )}
          </div>
        </div>
      </header>
    </>
  );
}

export default AppHeader;
