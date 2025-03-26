export const openLink = (url: string) => {
  // opens the link in a new tab
  window.setTimeout(() => window.open(url, "_blank"));
};
