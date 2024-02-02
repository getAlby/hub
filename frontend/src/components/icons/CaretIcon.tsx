import { SVGAttributes } from "react";

export function CaretIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      stroke="currentColor"
      strokeWidth="1"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        fillRule="evenodd"
        d="M12.248 8.237a.25.25 0 00-.354 0L5.53 14.601a.75.75 0 11-1.06-1.06l6.363-6.364a1.75 1.75 0 012.475 0l6.364 6.364a.75.75 0 01-1.06 1.06l-6.364-6.364z"
        clipRule="evenodd"
      />
    </svg>
  );
}
