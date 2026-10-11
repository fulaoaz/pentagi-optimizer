import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

import { dictionaries, fallbackLocale, translate } from '@/lib/i18n';

import { providerTestNameKeys, translateProviderTestName, translateProviderTestType } from './provider-test-labels';

const t = (key: string, values?: Record<string, number | string>) =>
    translate(dictionaries['zh-CN'], dictionaries[fallbackLocale], key, values);

describe('provider test labels', () => {
    it('covers every provider test declared by the backend registry', () => {
        const registryPath = resolve(process.cwd(), '../backend/pkg/providers/tester/testdata/tests.yml');
        const registry = readFileSync(registryPath, 'utf8');
        const names = [...registry.matchAll(/^ {2}name: "([^"]+)"$/gm)].map((match) => match[1]).sort();

        expect(Object.keys(providerTestNameKeys).sort()).toEqual(names);
    });

    it('translates fixed labels while preserving technical error details', () => {
        expect(translateProviderTestName('Simple Math Streaming', t)).toBe('简单数学运算（流式）');
        expect(translateProviderTestName('Execution Error: connection refused', t)).toBe(
            '执行错误：connection refused',
        );
        expect(translateProviderTestName('Future Upstream Test', t)).toBe('Future Upstream Test');
        expect(translateProviderTestType('completion', t)).toBe('文本生成');
        expect(translateProviderTestType('json', t)).toBe('JSON');
        expect(translateProviderTestType('custom', t)).toBe('custom');
    });
});
