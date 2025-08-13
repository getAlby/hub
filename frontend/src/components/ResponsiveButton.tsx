import { VariantProps } from "class-variance-authority";
import { LucideIcon } from "lucide-react";
import * as React from "react";
import { Button } from "./ui/button";
import { buttonVariants } from "./ui/buttonVariants";

type Props = {
  icon: LucideIcon;
  text: string;
} & React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

const ResponsiveButton = ({ icon: Icon, text, variant, ...props }: Props) => {
  return (
    <>
      <Button {...props} className="hidden lg:inline-flex" variant={variant}>
        <Icon />
        {text}
      </Button>
      <Button {...props} size="icon" className="lg:hidden" variant={variant}>
        <Icon />
      </Button>
    </>
  );
};

export default ResponsiveButton;
