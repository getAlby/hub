import type { FallbackLng } from "i18next";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "src/components/ui/select";
import { toast } from "src/components/ui/use-toast";
import i18n, { supportedLocales } from "src/i18n/i18nConfig";

export default function LocaleSwitcher() {
  const { t } = useTranslation("components", {
    keyPrefix: "locale_switcher",
  });
  const fallbackLng = i18n.options.fallbackLng?.[0 as keyof FallbackLng];
  const [dropdownLang, setDropdownLang] = useState(
    i18n.language || fallbackLng
  );

  const onLanguageChange = async (newLanguage: string) => {
    if (dropdownLang !== newLanguage) {
      setDropdownLang(newLanguage);
      i18n.changeLanguage(newLanguage);
      toast({ title: t("success") });
    }
  };

  return (
    <Select value={dropdownLang} onValueChange={onLanguageChange}>
      <SelectTrigger className="w-[150px]">
        <SelectValue placeholder="Language" />
      </SelectTrigger>
      <SelectContent>
        {supportedLocales.map((locale) => (
          <SelectItem key={locale.locale} value={locale.locale}>
            {locale.label}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
