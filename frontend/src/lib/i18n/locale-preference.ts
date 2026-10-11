import type { Locale } from './types';

import { defaultLocale, isLocale } from './types';

export const LOCALE_STORAGE_KEY = 'locale';

export const resolveLocalePreference = (stored: unknown, _preferredLanguages: readonly string[]): Locale => {
    if (isLocale(stored)) {
        return stored;
    }

    return defaultLocale;
};

export const resolveBrowserLocale = (storageKey = LOCALE_STORAGE_KEY): Locale =>
    resolveLocalePreference(localStorage.getItem(storageKey), globalThis.navigator?.languages ?? []);
