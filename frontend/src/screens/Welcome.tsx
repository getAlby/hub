import { Link } from "react-router-dom";

import nwcComboMark from "src/assets/images/nwc-combomark.svg";

export function Welcome() {
  return (
    <div className="bg-white rounded-md shadow p-4 lg:p-8 mb-8 mx-auto max-w-xl dark:bg-surface-02dp flex flex-col items-center">
      <div className="px-8 py-4 mb-14">
        <img alt="NWC Logo" className="h-12 inline" src={nwcComboMark} />
      </div>

      <h1 className="font-semibold text-2xl font-headline mb-2 dark:text-white">
        Welcome to NWC app
      </h1>
      <p className="text-center font-light text-md leading-relaxed dark:text-neutral-400 px-4 mb-14">
        Connect your lightning wallet to dozens of apps and enjoy seamless
        in-app payments.
      </p>

      <div className="w-full flex flex-row justify-between">
        <Link
          to="/about"
          className="flex-row w-1/2 first:mr-2 last:ml-2 px-0 py-2 bg-white text-gray-700 dark:bg-surface-02dp dark:text-neutral-200 dark:border-neutral-800 hover:bg-gray-50 dark:hover:bg-surface-16dp cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-700 transition duration-150"
        >
          Learn more
        </Link>
        <Link
          to="/setup"
          className="flex-row w-1/2 first:mr-2 last:ml-2 px-0 py-2 bg-purple-700 border-2 border-transparent text-white hover:bg-purple-800 cursor-pointer inline-flex justify-center items-center font-medium bg-origin-border shadow rounded-md focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-purple-500 transition duration-150"
        >
          Get Started
        </Link>
      </div>
    </div>
  );
}
