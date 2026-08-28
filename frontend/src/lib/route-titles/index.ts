import type { ComponentType } from 'react';

import type { Translate } from '@/lib/i18n';

import {
    useFlowQuery,
    useFlowTemplateQuery,
    useKnowledgeDocumentQuery,
    useSettingsProvidersQuery,
} from '@/graphql/types';

import { apolloTitle } from './apollo-title';
import { formatPromptId } from './format-prompt-id';
import { type RouteParams } from './render-title';

export interface RouteTitleHandle {
    title: TitleResolver;
}

/**
 * A `handle.title` value can be one of three forms:
 *   - `string` — a translation key, resolved by `DocumentTitle` via `t()`.
 *   - `(params, t) => string` — derived synchronously from URL params.
 *   - `ComponentType<{ params }>` — reactive (e.g. subscribes to Apollo
 *     cache for resource-driven titles). Must be produced by `apolloTitle()`
 *     so the marker it attaches lets `DocumentTitle` distinguish a component
 *     from a `(params, t) => string` resolver at runtime. A hand-rolled
 *     component function will be misdetected as a resolver and called with raw
 *     params — always route reactive titles through `apolloTitle()`.
 */
export type TitleResolver =
    | ((params: RouteParams, t: Translate) => string)
    | ComponentType<{ params: RouteParams }>
    | string;

/**
 * Single source of truth for every route's document `<title>`. `app.tsx`
 * imports nothing from Apollo for title purposes — it only spreads handles
 * from this registry onto the matching <Route>.
 */
export const routeTitles = {
    apiTokens: { title: 'title.apiTokens' },
    dashboard: { title: 'title.dashboard' },
    flow: {
        title: apolloTitle({
            select: (data, { flowId }, t) =>
                data?.flow?.title && flowId
                    ? t('title.flowNumbered', { id: flowId, title: data.flow.title })
                    : t('title.flow'),
            useQuery: useFlowQuery,
            variables: ({ flowId }) => (flowId ? { id: flowId } : null),
        }),
    },
    flowReport: { title: 'title.flowReport' },
    flows: { title: 'title.flows' },
    knowledge: {
        title: apolloTitle({
            select: (data, { knowledgeId }, t) =>
                knowledgeId === 'new'
                    ? t('title.newKnowledge')
                    : data?.knowledgeDocument?.question || t('title.knowledge'),
            useQuery: useKnowledgeDocumentQuery,
            variables: ({ knowledgeId }) => (!knowledgeId || knowledgeId === 'new' ? null : { id: knowledgeId }),
        }),
    },
    knowledges: { title: 'title.knowledges' },
    login: { title: 'title.login' },
    newFlow: { title: 'title.newFlow' },
    oauth: { title: 'title.oauth' },
    prompt: {
        title: (params: RouteParams, t: Translate) =>
            params.promptId ? formatPromptId(params.promptId, t) : t('title.prompt'),
    },
    prompts: { title: 'title.prompts' },

    provider: {
        title: apolloTitle({
            select: (data, { providerId }, t) => {
                if (providerId === 'new') {
                    return t('title.newProvider');
                }

                const provider = data?.settingsProviders.userDefined?.find(
                    (candidate) => String(candidate.id) === providerId,
                );

                return provider?.name || t('title.provider');
            },
            useQuery: useSettingsProvidersQuery,
            variables: ({ providerId }) => (providerId === 'new' ? null : {}),
        }),
    },

    providers: { title: 'title.providers' },

    resources: { title: 'title.resources' },

    template: {
        title: apolloTitle({
            select: (data, { templateId }, t) =>
                templateId === 'new' ? t('title.newTemplate') : data?.flowTemplate?.title || t('title.template'),
            useQuery: useFlowTemplateQuery,
            variables: ({ templateId }) => (!templateId || templateId === 'new' ? null : { templateId }),
        }),
    },

    templates: { title: 'title.templates' },
} as const satisfies Record<string, RouteTitleHandle>;
