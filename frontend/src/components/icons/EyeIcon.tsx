import { SVGAttributes } from "react";

export function EyeIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M12 14a2 2 0 100-4 2 2 0 000 4z" />
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M21 12c0 2.761-4.03 5-9 5s-9-2.239-9-5 4.03-5 9-5 9 2.239 9 5zm-5 0a4 4 0 11-8 0 4 4 0 018 0z"
      />
    </svg>
  );
}
