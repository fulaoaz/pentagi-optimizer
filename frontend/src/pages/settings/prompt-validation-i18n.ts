import type { PromptValidationResultFragmentFragment } from '@/graphql/types';
import type { Translate } from '@/lib/i18n';

import { PromptValidationErrorType } from '@/graphql/types';

interface PromptValidationCopy {
    details: null | string;
    message: null | string;
}

const detailTranslationKeys: Record<string, string> = {
    'Array/slice index out of bounds or incorrect index type': 'settings.prompts.validation.indexError',
    "Check for missing closing braces '}}' or incorrect template syntax":
        'settings.prompts.validation.checkClosingBraces',
    'Check variable types and data structure in template': 'settings.prompts.validation.checkDataStructure',
    'Referenced variable or field not found in provided data': 'settings.prompts.validation.variableNotFound',
    'Review template syntax according to Go template documentation':
        'settings.prompts.validation.reviewGoTemplateSyntax',
    'Template appears to be incomplete - missing closing braces': 'settings.prompts.validation.incompleteTemplate',
    'These variables are not declared in PromptVariables for this prompt type. Backend code cannot provide these variables.':
        'settings.prompts.validation.unauthorizedVariablesDetail',
    'Unknown function or incorrect function call syntax': 'settings.prompts.validation.unknownFunction',
    'Variable type mismatch - check if template expects different data structure':
        'settings.prompts.validation.variableTypeMismatchDetail',
};

const stripPrefix = (message: null | string, prefix: string): null | string => {
    const value = message?.trim();

    if (!value) {
        return null;
    }

    return value.startsWith(prefix) ? value.slice(prefix.length).trim() || null : value;
};

const localizeDetails = (details: null | string, t: Translate): null | string => {
    if (!details) {
        return null;
    }

    const key = detailTranslationKeys[details];

    return key ? t(key) : details;
};

export const getPromptValidationCopy = (
    result: PromptValidationResultFragmentFragment,
    t: Translate,
): PromptValidationCopy => {
    const details = localizeDetails(result.details, t);

    switch (result.errorType) {
        case PromptValidationErrorType.EmptyTemplate:
            return { details, message: t('settings.prompts.validation.emptyTemplate') };

        case PromptValidationErrorType.RenderingFailed: {
            const reason = stripPrefix(result.message, 'template rendering failed:');

            return {
                details,
                message: reason
                    ? t('settings.prompts.validation.renderingFailedWithReason', { reason })
                    : t('settings.prompts.validation.renderingFailed'),
            };
        }

        case PromptValidationErrorType.SyntaxError: {
            const reason = stripPrefix(result.message, 'failed to parse template:');

            return {
                details,
                message: reason
                    ? t('settings.prompts.validation.syntaxErrorWithReason', { reason })
                    : t('settings.prompts.validation.syntaxError'),
            };
        }

        case PromptValidationErrorType.UnauthorizedVariable: {
            const promptType = stripPrefix(result.message, 'unknown prompt type:');

            if (result.message?.startsWith('unknown prompt type:') && promptType) {
                return {
                    details,
                    message: t('settings.prompts.validation.unknownPromptType', { promptType }),
                };
            }

            const variables = result.message?.match(/\[[^\]]*\]/)?.[0];

            return {
                details,
                message: variables
                    ? t('settings.prompts.validation.unauthorizedVariables', { variables })
                    : t('settings.prompts.validation.unauthorizedVariable'),
            };
        }

        case PromptValidationErrorType.VariableTypeMismatch:
            return { details, message: t('settings.prompts.validation.variableTypeMismatch') };

        case PromptValidationErrorType.UnknownType:
        default:
            return { details, message: t('settings.prompts.validation.unknownError') };
    }
};
