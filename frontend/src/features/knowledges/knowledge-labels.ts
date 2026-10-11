import type { Translate } from '@/lib/i18n';
import type { Knowledge } from '@/providers/knowledges-provider';

import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType } from '@/graphql/types';

const docTypeKeys: Record<KnowledgeDocType, string> = {
    [KnowledgeDocType.Answer]: 'knowledge.docType.answer',
    [KnowledgeDocType.Code]: 'knowledge.docType.code',
    [KnowledgeDocType.Guide]: 'knowledge.docType.guide',
};

const answerTypeKeys: Record<KnowledgeAnswerType, string> = {
    [KnowledgeAnswerType.Code]: 'knowledge.answerType.code',
    [KnowledgeAnswerType.Guide]: 'knowledge.answerType.guide',
    [KnowledgeAnswerType.Other]: 'knowledge.answerType.other',
    [KnowledgeAnswerType.Tool]: 'knowledge.answerType.tool',
    [KnowledgeAnswerType.Vulnerability]: 'knowledge.answerType.vulnerability',
};

const guideTypeKeys: Record<KnowledgeGuideType, string> = {
    [KnowledgeGuideType.Configure]: 'knowledge.guideType.configure',
    [KnowledgeGuideType.Development]: 'knowledge.guideType.development',
    [KnowledgeGuideType.Install]: 'knowledge.guideType.install',
    [KnowledgeGuideType.Other]: 'knowledge.guideType.other',
    [KnowledgeGuideType.Pentest]: 'knowledge.guideType.pentest',
    [KnowledgeGuideType.Use]: 'knowledge.guideType.use',
};

export const getKnowledgeDocTypeLabel = (t: Translate, value: KnowledgeDocType): string => t(docTypeKeys[value]);

export const getKnowledgeAnswerTypeLabel = (t: Translate, value: KnowledgeAnswerType): string =>
    t(answerTypeKeys[value]);

export const getKnowledgeGuideTypeLabel = (t: Translate, value: KnowledgeGuideType): string => t(guideTypeKeys[value]);

export const getKnowledgeSubtypeLabel = (t: Translate, knowledge: Knowledge): null | string => {
    if (knowledge.docType === KnowledgeDocType.Guide && knowledge.guideType) {
        return getKnowledgeGuideTypeLabel(t, knowledge.guideType);
    }

    if (knowledge.docType === KnowledgeDocType.Answer && knowledge.answerType) {
        return getKnowledgeAnswerTypeLabel(t, knowledge.answerType);
    }

    if (knowledge.docType === KnowledgeDocType.Code) {
        return knowledge.codeLang ?? null;
    }

    return null;
};
