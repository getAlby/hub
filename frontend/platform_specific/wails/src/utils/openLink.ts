import { BrowserOpenURL } from "wailsjs/runtime/runtime";

export const openLink = (url: string) => {
  // opens the link in the browser
  try {
    BrowserOpenURL(url);
  } catch (error) {
    console.error("Failed to open link", error);
    throw error;
  }
};
