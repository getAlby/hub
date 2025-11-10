import { VariantProps } from "class-variance-authority";
import * as React from "react";
import { LinkButton } from "src/components/ui/custom/link-button";
import { buttonVariants } from "./ui/buttonVariants";

type Props = {
  icon: React.ComponentType;
  text: string;
  to: string;
} & React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  };

const ResponsiveLinkButton = ({
  icon: Icon,
  text,
  variant,
  ...props
}: Props) => {
  return (
    <>
      <LinkButton
        {...props}
        className="hidden lg:inline-flex"
        variant={variant}
      >
        <Icon />
        {text}
      </LinkButton>
      <LinkButton
        {...props}
        size="icon"
        className="lg:hidden"
        variant={variant}
      >
        <Icon />
      </LinkButton>
    </>
  );
};

export default ResponsiveLinkButton;
