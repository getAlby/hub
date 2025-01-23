import React from "react";
import { Separator } from "src/components/ui/separator";

type Props = {
  title: string;
  description: string | React.ReactNode;
};

function SettingsHeader({ title, description }: Props) {
  return (
    <>
      <div className="space-y-6">
        <div>
          <h3 className="text-2xl font-medium">{title}</h3>
          <p className="text-base text-muted-foreground">{description}</p>
        </div>
        <Separator />
      </div>
    </>
  );
}

export default SettingsHeader;
