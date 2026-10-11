import { describe, expect, it } from 'vitest';

import type { Translate } from '@/lib/i18n';

import { dictionaries, translate } from '@/lib/i18n';

import { buildFlowFilesPromoteFormSchema } from './use-flow-files-promote';

const t: Translate = (key, values) => translate(dictionaries['zh-CN'], dictionaries.en, key, values);

describe('buildFlowFilesPromoteFormSchema', () => {
    it('allows the resource-library root for a multi-item save', () => {
        expect(buildFlowFilesPromoteFormSchema(t, true).safeParse({ destination: '' }).success).toBe(true);
    });

    it('requires an explicit destination for a single item', () => {
        const result = buildFlowFilesPromoteFormSchema(t).safeParse({ destination: '' });

        expect(result.success).toBe(false);
        expect(result.error?.issues[0]?.message).toBe('请输入保存位置');
    });
});
