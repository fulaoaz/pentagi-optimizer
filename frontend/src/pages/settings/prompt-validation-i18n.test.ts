import { describe, expect, it } from 'vitest';

import type { PromptValidationResultFragmentFragment } from '@/graphql/types';

import { PromptValidationErrorType, ResultType } from '@/graphql/types';
import { dictionaries, fallbackLocale, translate } from '@/lib/i18n';

import { getPromptValidationCopy } from './prompt-validation-i18n';

const t = (key: string, values?: Record<string, number | string>) =>
    translate(dictionaries['zh-CN'], dictionaries[fallbackLocale], key, values);

const makeResult = (
    values: Partial<PromptValidationResultFragmentFragment>,
): PromptValidationResultFragmentFragment => ({
    details: null,
    errorType: PromptValidationErrorType.UnknownType,
    line: null,
    message: null,
    result: ResultType.Error,
    ...values,
});

describe('getPromptValidationCopy', () => {
    it('localizes an empty template message', () => {
        const copy = getPromptValidationCopy(
            makeResult({
                errorType: PromptValidationErrorType.EmptyTemplate,
                message: 'template content cannot be empty',
            }),
            t,
        );

        expect(copy.message).toBe('模板内容不能为空。');
    });

    it('preserves variable identifiers in an unauthorized-variable message', () => {
        const copy = getPromptValidationCopy(
            makeResult({
                details:
                    'These variables are not declared in PromptVariables for this prompt type. Backend code cannot provide these variables.',
                errorType: PromptValidationErrorType.UnauthorizedVariable,
                message: 'template uses unauthorized variables: [TargetURL UserName]',
            }),
            t,
        );

        expect(copy.message).toBe('模板使用了未授权变量：[TargetURL UserName]。');
        expect(copy.details).toBe('这些变量未在当前提示词类型的 PromptVariables 中声明，后端不会提供相应值。');
    });

    it('localizes the explanation while preserving parser diagnostics', () => {
        const copy = getPromptValidationCopy(
            makeResult({
                details: "Check for missing closing braces '}}' or incorrect template syntax",
                errorType: PromptValidationErrorType.SyntaxError,
                message: 'failed to parse template: template:1: unexpected EOF',
            }),
            t,
        );

        expect(copy.message).toBe('模板语法解析失败：template:1: unexpected EOF');
        expect(copy.details).toBe('请检查是否缺少右花括号“}}”，或模板语法是否正确。');
    });
});
