import { SVGAttributes } from "react";

export function TransactionsIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <circle cx="5.5" cy="7.5" r="1.5" />
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M8 6.5a.5.5 0 01.5-.5h11a.5.5 0 010 1h-11a.5.5 0 01-.5-.5zM8 8.5a.5.5 0 01.5-.5h6a.5.5 0 010 1h-6a.5.5 0 01-.5-.5z"
      />
      <circle cx="5.5" cy="12" r="1.5" />
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M8 11a.5.5 0 01.5-.5h8a.5.5 0 010 1h-8A.5.5 0 018 11zM8 13a.5.5 0 01.5-.5h7a.5.5 0 010 1h-7A.5.5 0 018 13z"
      />
      <circle cx="5.5" cy="16.5" r="1.5" />
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M8 15.5a.5.5 0 01.5-.5H18a.5.5 0 010 1H8.5a.5.5 0 01-.5-.5zM8 17.5a.5.5 0 01.5-.5h4a.5.5 0 010 1h-4a.5.5 0 01-.5-.5z"
      />
    </svg>
  );
}
