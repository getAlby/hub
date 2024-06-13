import { ReactElement } from "react";
import Breadcrumbs from "src/components/Breadcrumbs";

type Props = {
  title: string | ReactElement;
  description: string | ReactElement;
  contentRight?: React.ReactNode;
};

function AppHeader({ title, description, contentRight }: Props) {
  return (
    <>
      <Breadcrumbs />
      <div className="flex justify-between border-b border-border pb-3 lg:pb-6">
        <div className="flex-1">
          <h1 className="text-xl lg:text-3xl font-semibold">{title}</h1>
          <p className="hidden lg:inline text-muted-foreground">
            {description}
          </p>
        </div>
        <div className="flex gap-3">{contentRight}</div>
      </div>
    </>
  );
}

export default AppHeader;
