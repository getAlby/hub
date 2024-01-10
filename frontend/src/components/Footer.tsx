import { useDarkMode } from "@hooks/useDarkMode";
import albyLogoWithText from "@assets/images/alby-logo-with-text.svg";
import albyLogoWithTextDark from "@assets/images/alby-logo-with-text-dark.svg";

function Footer() {
  const isDarkMode = useDarkMode();
  return (
    <footer className="mt-20 mb-4 flex justify-center items-center gap-2">
      <span className="text-gray-500 dark:text-neutral-300 text-xs">
        Made with 💜 by
      </span>
      <a href="https://getalby.com?utm_source=nwc" rel="noreferrer noopener">
        <img
          id="alby-logo"
          src={isDarkMode ? albyLogoWithTextDark : albyLogoWithText}
          className="w-16 inline"
        />
      </a>
    </footer>
  );
}

export default Footer;
