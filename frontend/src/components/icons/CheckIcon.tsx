import { SVGAttributes } from "react";

export function CheckIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path
        fill-rule="evenodd"
        clip-rule="evenodd"
        d="M18.3814 5.35432C18.738 5.56495 18.8564 6.02484 18.6458 6.3815L11.5591 18.3815C11.4426 18.5787 11.2424 18.7119 11.0156 18.7431C10.7887 18.7743 10.56 18.7002 10.3946 18.5418L5.48126 13.8366C5.18211 13.5501 5.17184 13.0753 5.45833 12.7761C5.74482 12.477 6.21958 12.4667 6.51874 12.7532L10.7487 16.804L17.3542 5.61874C17.5648 5.26208 18.0247 5.14369 18.3814 5.35432Z"
        fill="black"
      />
    </svg>
  );
}
