import type { ReactNode } from 'react';

import { createContext, useCallback, useEffect, useMemo, useState } from 'react';

import type { Locale, TranslationValues } from '@/lib/i18n';

import {
    defaultLocale,
    dictionaries,
    fallbackLocale,
    LOCALE_STORAGE_KEY,
    resolveBrowserLocale,
    translate,
} from '@/lib/i18n';

interface LocaleProviderState {
    locale: Locale;
    setLocale: (locale: Locale) => void;
    /** Translates `key`, interpolating `{name}` placeholders from `values`. */
    t: (key: string, values?: TranslationValues) => string;
}

export const LocaleProviderContext = createContext<LocaleProviderState>({
    locale: defaultLocale,
    setLocale: () => null,
    t: (key, values) => translate(dictionaries[fallbackLocale], dictionaries[fallbackLocale], key, values),
});

interface LocaleProviderProps {
    children: ReactNode;
    storageKey?: string;
}

export function LocaleProvider({ children, storageKey = LOCALE_STORAGE_KEY }: LocaleProviderProps) {
    const [locale, setLocaleState] = useState<Locale>(() => resolveBrowserLocale(storageKey));

    useEffect(() => {
        document.documentElement.lang = locale;
    }, [locale]);

    const setLocale = useCallback(
        (next: Locale) => {
            localStorage.setItem(storageKey, next);
            setLocaleState(next);
        },
        [storageKey],
    );

    const value = useMemo<LocaleProviderState>(
        () => ({
            locale,
            setLocale,
            t: (key, values) => translate(dictionaries[locale], dictionaries[fallbackLocale], key, values),
        }),
        [locale, setLocale],
    );

    return <LocaleProviderContext.Provider value={value}>{children}</LocaleProviderContext.Provider>;
}
