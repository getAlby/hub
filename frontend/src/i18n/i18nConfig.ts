// dayjs locales must be imported as well, list: https://github.com/iamkun/dayjs/tree/dev/src/locale
import "dayjs/locale/hi";
import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import { initReactI18next } from "react-i18next";
// import our translations
import en from "src/i18n/locales/en/translation.json";
import hi from "src/i18n/locales/hi/translation.json";

export const defaultNS = "translation";
// needs to be aligned with `supportedLocales`
export const resources = {
  en: {
    translation: en.translation,
    common: en.common,
    components: en.components,
  },
  hi: {
    translation: hi.translation,
    common: hi.common,
    components: hi.components,
  },
} as const;

// needs to be aligned with `resources`
export const supportedLocales = [
  { locale: "en", label: "English" },
  { locale: "hi", label: "हिंदी" },
];

i18n
  // pass the i18n instance to react-i18next.
  .use(initReactI18next)
  // detect user language
  // learn more: https://github.com/i18next/i18next-browser-languageDetector
  .use(LanguageDetector)
  // init i18next
  // for all options read: https://www.i18next.com/overview/configuration-options
  .init({
    fallbackLng: "en",
    ns: ["translation", "common", "components", "permissions"],
    defaultNS,
    resources,
    supportedLngs: supportedLocales.map(({ locale }) => locale),
  });

export default i18n;
