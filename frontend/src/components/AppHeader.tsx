import { ReactElement } from "react";
import { Separator } from "src/components/ui/separator";
import { SidebarTrigger } from "src/components/ui/sidebar";
import { cn } from "src/lib/utils";

type Props = {
  title: string | ReactElement;
  titleClassName?: string;
  description?: string | ReactElement;
  contentRight?: React.ReactNode;
  breadcrumb?: boolean;
  addSidebarTrigger?: boolean;
};

function AppHeader({
  title,
  titleClassName,
  description = "",
  contentRight,
  addSidebarTrigger = true,
}: Props) {
  return (
    <>
      <header className="flex flex-row items-center border-b border-border pb-4 gap-2">
        {addSidebarTrigger && <SidebarTrigger className="-ml-1 md:hidden" />}
        <Separator orientation="vertical" className="mr-2 h-4 md:hidden" />
        <div className="flex flex-col flex-1">
          <div className="flex justify-between items-center">
            <div className="flex-1">
              <h1
                className={cn(
                  "text-2xl lg:text-3xl font-semibold",
                  titleClassName
                )}
              >
                {title}
              </h1>
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
