import { describe, expect, it } from 'vitest';

import type { Translate } from '@/lib/i18n';

import { dictionaries, translate } from '@/lib/i18n';

import { formatMessageForClipboard } from './clipboard';

const zhCNTranslate: Translate = (key, values) => translate(dictionaries['zh-CN'], dictionaries.en, key, values);
const enTranslate: Translate = (key, values) => translate(dictionaries.en, dictionaries.en, key, values);

describe('formatMessageForClipboard', () => {
    it('uses the English section labels with the English translator', async () => {
        const content = await formatMessageForClipboard(
            {
                result: 'scan complete',
                thinking: 'inspect the target',
            },
            enTranslate,
        );

        expect(content).toContain('<summary>Thinking</summary>');
        expect(content).toContain('<summary>Result</summary>');
    });

    it('localizes section labels with the active translator', async () => {
        const content = await formatMessageForClipboard(
            {
                result: '扫描完成',
                thinking: '检查目标',
            },
            zhCNTranslate,
        );

        expect(content).toContain('<summary>思考过程</summary>');
        expect(content).toContain('<summary>结果</summary>');
    });
});
