import { SVGAttributes } from "react";

export function MastercardLogo(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      {...props}
      viewBox="0 0 152 108"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Mastercard"
    >
      <circle cx="60" cy="54" r="36" fill="#EB001B" />
      <circle cx="92" cy="54" r="36" fill="#F79E1B" />
      <path
        fill="#FF5F00"
        d="M 76 21.75 A 36 36 0 0 1 76 86.25 A 36 36 0 0 1 76 21.75 Z"
      />
    </svg>
  );
}
