import { SVGAttributes } from "react";

export function LogoutIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        d="M19.2854 12.0002L11.2727 12.0002"
        stroke="currentColor"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
      />
      <path
        d="M16.5101 15.6364L19.9999 12L16.5101 8.36363"
        stroke="currentColor"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
      />
      <path
        d="M13.4545 7V4H4V20H13.4545V17"
        stroke="currentColor"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="2"
      />
    </svg>
  );
}
