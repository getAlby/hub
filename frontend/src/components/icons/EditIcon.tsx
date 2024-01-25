import { SVGAttributes } from "react";

export function EditIcon(props: SVGAttributes<SVGElement>) {
  return (
    <svg
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="currentColor"
      xmlns="http://www.w3.org/2000/svg"
      {...props}
    >
      <path d="M3.995 17.207V19.5a.5.5 0 00.5.5h2.298a.5.5 0 00.353-.146l9.448-9.448-3-3-9.452 9.448a.5.5 0 00-.147.353zM14.832 6.167l3 3 1.46-1.46a1 1 0 000-1.414l-1.585-1.586a1 1 0 00-1.414 0l-1.46 1.46z" />
    </svg>
  );
}
