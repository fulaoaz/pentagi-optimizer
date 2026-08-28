import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import ts from 'typescript';
import { describe, expect, it } from 'vitest';

import { dictionaries, fallbackLocale, interpolate, translate } from './index';
import { locales } from './types';

const keysOf = (locale: (typeof locales)[number]) => Object.keys(dictionaries[locale]).sort();

describe('dictionaries', () => {
    it('does not declare duplicate keys in locale source files', () => {
        for (const locale of locales) {
            const source = readFileSync(resolve('src/lib/i18n/locales', `${locale}.ts`), 'utf8');
            const sourceFile = ts.createSourceFile(
                `${locale}.ts`,
                source,
                ts.ScriptTarget.Latest,
                true,
                ts.ScriptKind.TS,
            );
            const declaredKeys = new Set<string>();
            const duplicates = new Set<string>();

            const visit = (node: ts.Node) => {
                if (ts.isPropertyAssignment(node) && ts.isStringLiteral(node.name)) {
                    if (declaredKeys.has(node.name.text)) {
                        duplicates.add(node.name.text);
                    }

                    declaredKeys.add(node.name.text);
                }

                ts.forEachChild(node, visit);
            };

            visit(sourceFile);
            expect([...duplicates], `${locale} contains duplicate dictionary keys`).toEqual([]);
        }
    });

    it('ships a dictionary for every declared locale', () => {
        for (const locale of locales) {
            expect(Object.keys(dictionaries[locale]).length).toBeGreaterThan(0);
        }
    });

    it('keeps every locale in sync with the fallback key set', () => {
        const expected = keysOf(fallbackLocale);

        for (const locale of locales) {
            expect(keysOf(locale), `locale "${locale}" key set drifted from "${fallbackLocale}"`).toEqual(expected);
        }
    });

    it('has no empty message values', () => {
        for (const locale of locales) {
            for (const [key, value] of Object.entries(dictionaries[locale])) {
                expect(value.trim(), `${locale}:${key} is empty`).not.toBe('');
            }
        }
    });

    it('uses the same placeholders across locales', () => {
        const placeholders = (template: string) => [...template.matchAll(/\{(\w+)\}/g)].map((match) => match[1]).sort();

        for (const [key, template] of Object.entries(dictionaries[fallbackLocale])) {
            for (const locale of locales) {
                // The key-set test above guarantees the lookup is present; `?? ''`
                // only satisfies `noUncheckedIndexedAccess`.
                expect(placeholders(dictionaries[locale][key] ?? ''), `${locale}:${key} placeholder mismatch`).toEqual(
                    placeholders(template),
                );
            }
        }
    });
});

describe('interpolate', () => {
    it('substitutes provided values', () => {
        expect(interpolate('Flow #{id} — {title}', { id: 7, title: 'scan' })).toBe('Flow #7 — scan');
    });

    it('leaves unknown placeholders verbatim', () => {
        expect(interpolate('{a} and {b}', { a: 'x' })).toBe('x and {b}');
    });

    it('returns the template unchanged when no values are given', () => {
        expect(interpolate('{a}')).toBe('{a}');
    });
});

describe('translate', () => {
    it('prefers the active dictionary', () => {
        expect(translate({ 'common.save': '保存' }, { 'common.save': 'Save' }, 'common.save')).toBe('保存');
    });

    it('falls back when the key is untranslated', () => {
        expect(translate({}, { 'common.save': 'Save' }, 'common.save')).toBe('Save');
    });

    it('returns the key when it is missing everywhere', () => {
        expect(translate({}, {}, 'nope.missing')).toBe('nope.missing');
    });
});
