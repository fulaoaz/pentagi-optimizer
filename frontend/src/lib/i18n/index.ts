import type { Dictionary, Locale } from './types';

import { en } from './locales/en';
import { zhCN } from './locales/zh-CN';

export { LOCALE_STORAGE_KEY, resolveBrowserLocale, resolveLocalePreference } from './locale-preference';
export { interpolate, translate } from './translate';
export {
    defaultLocale,
    type Dictionary,
    isLocale,
    type Locale,
    localeNames,
    locales,
    type Translate,
    type TranslationValues,
} from './types';

/**
 * Locale used to resolve keys the active dictionary is missing. English is the
 * upstream source of truth, so a key added upstream but not yet translated
 * still renders real text instead of a raw key.
 */
export const fallbackLocale: Locale = 'en';

export const dictionaries: Record<Locale, Dictionary> = {
    en: en,
    'zh-CN': zhCN,
};
