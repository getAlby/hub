import { SVGAttributes } from "react";

export function CopyIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <rect width="10" height="14" x="6" y="6" rx="1.5" />
      <path d="M8.064 5.064A1.5 1.5 0 018.5 5h7A1.5 1.5 0 0117 6.5v11a1.5 1.5 0 01-.064.436A1.5 1.5 0 0018 16.5v-11A1.5 1.5 0 0016.5 4h-7a1.5 1.5 0 00-1.436 1.064z" />
    </svg>
  );
}
