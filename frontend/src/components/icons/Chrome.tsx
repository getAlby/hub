import { SVGAttributes } from "react";

export function ChromeIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="16"
      height="16"
      fill="none"
      {...props}
    >
      <path
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M8.07 5.291h6.505M1 8.071a7.07 7.07 0 1 0 14.141 0A7.07 7.07 0 0 0 1 8.07Z"
      />
      <path
        stroke="currentColor"
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M5.651 9.461 2.398 3.828m8.07 5.628-3.252 5.633m-1.928-7.02a2.778 2.778 0 1 0 5.556 0 2.778 2.778 0 0 0-5.556 0Z"
      />
    </svg>
  );
}
