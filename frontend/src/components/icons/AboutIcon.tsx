import { SVGAttributes } from "react";

export function AboutIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <circle cx="12" cy="12" r="8.5" />
      <path strokeLinecap="round" d="M12 10.5v7M12 8V7"></path>
    </svg>
  );
}
