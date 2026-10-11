import type { Dictionary, TranslationValues } from './types';

/**
 * Replaces `{name}` placeholders with values. Unknown placeholders are left
 * verbatim so a missing value is visible in the UI rather than silently empty.
 */
export const interpolate = (template: string, values?: TranslationValues): string => {
    if (!values) {
        return template;
    }

    return template.replaceAll(/\{(\w+)\}/g, (match, name: string) => (name in values ? String(values[name]) : match));
};

/**
 * Resolves `key` against `dictionary`, falling back to `fallback` and finally
 * to the key itself. Returning the key (instead of throwing) keeps an
 * untranslated string from taking down the whole screen.
 */
export const translate = (
    dictionary: Dictionary,
    fallback: Dictionary,
    key: string,
    values?: TranslationValues,
): string => {
    const template = dictionary[key] ?? fallback[key];

    if (template === undefined) {
        if (import.meta.env.DEV) {
            console.warn(`[i18n] missing translation key: ${key}`);
        }

        return key;
    }

    return interpolate(template, values);
};
