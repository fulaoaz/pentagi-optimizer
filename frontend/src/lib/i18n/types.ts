export const locales = ['en', 'zh-CN'] as const;

export type Locale = (typeof locales)[number];

/**
 * This maintained fork is Chinese-first. The upstream English dictionary is
 * retained as a selectable locale rather than being replaced or discarded.
 */
export const defaultLocale: Locale = 'zh-CN';

export const localeNames: Record<Locale, string> = {
    en: 'English',
    'zh-CN': '简体中文',
};

export const isLocale = (value: unknown): value is Locale =>
    typeof value === 'string' && locales.includes(value as Locale);

/**
 * A translation dictionary is a flat map of dot-separated keys to strings.
 * Flat (rather than nested) keeps lookup a single property access and makes
 * missing-key diffs between locales trivial to compute in tests.
 */
export type Dictionary = Record<string, string>;

/**
 * The translator returned by `useLocale()`. Exported so helpers that build
 * localized values outside a component (zod schemas, column definitions) can
 * take it as a parameter.
 */
export type Translate = (key: string, values?: TranslationValues) => string;

/** Values interpolated into a message via `{name}` placeholders. */
export type TranslationValues = Record<string, number | string>;
