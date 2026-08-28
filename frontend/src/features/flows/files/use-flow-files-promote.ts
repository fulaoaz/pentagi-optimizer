import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import type { OverwriteOutcome } from '@/components/shared/overwrite';
import type { RestResourceList } from '@/features/resources/resources-rest';
import type { Translate } from '@/lib/i18n';

import { useLocale } from '@/hooks/use-locale';
import { api, getApiErrorMessage, getApiErrorStatusCode } from '@/lib/axios';

import { FLOW_FILES_PROMOTE_API_PATH } from './flow-files-constants';

export const buildFlowFilesPromoteFormSchema = (t: Translate, allowLibraryRoot = false) =>
    z.object({
        destination: z
            .string()
            .trim()
            .refine((value) => allowLibraryRoot || value.length > 0, {
                message: t('flow.files.destinationCannotBeEmpty'),
            })
            .refine((value) => !value.startsWith('/'), { message: t('flow.files.destinationRelative') })
            .refine((value) => !value.split('/').includes('..'), {
                message: t('flow.files.destinationNoParentSegment'),
            }),
    });

export type FlowFilesPromoteFormValues = z.infer<ReturnType<typeof buildFlowFilesPromoteFormSchema>>;

interface PromoteRequestBody {
    destination: string;
    force: boolean;
    sources: readonly string[];
}

interface UseFlowFilesPromoteParams {
    flowId: null | string;
}

interface UseFlowFilesPromoteResult {
    isPromoting: boolean;
    /**
     * Issue a batch promote in a single atomic request and return a discriminated outcome:
     *   - `ok`        — every flow file/dir was promoted (success toast already fired),
     *   - `conflict`  — at least one resource path is occupied (no toast, caller
     *                   resolves via the shared overwrite workflow),
     *   - `error`     — anything else (failure toast already fired).
     *
     * Backend semantics: with one source, `destination` is the exact target
     * path; with multiple sources, `destination` is a base directory and each
     * source lands at `destination/<basename>`. The Apollo cache stays in sync
     * via `resourceAdded` / `resourceUpdated` GraphQL subscriptions.
     */
    promote: (sources: readonly string[], destination: string, force: boolean) => Promise<OverwriteOutcome>;
}

/**
 * Wraps the "promote flow file → user resource" REST call (`POST /files/to-resources`)
 * with toast notifications and a loading flag.
 */
export function useFlowFilesPromote({ flowId }: UseFlowFilesPromoteParams): UseFlowFilesPromoteResult {
    const { t } = useLocale();
    const [isPromoting, setIsPromoting] = useState(false);

    const promote = useCallback(
        async (sources: readonly string[], destination: string, force: boolean): Promise<OverwriteOutcome> => {
            if (!flowId || sources.length === 0) {
                return { kind: 'error' };
            }

            setIsPromoting(true);

            try {
                await api.post<RestResourceList, PromoteRequestBody>(
                    FLOW_FILES_PROMOTE_API_PATH(flowId),
                    {
                        destination: destination.trim(),
                        force,
                        sources,
                    },
                    { timeout: 0 },
                );

                const count = t('fileManager.itemCountMany', { count: sources.length });
                const normalizedDestination = destination.trim();
                let description: string;

                if (sources.length === 1) {
                    description = t('flow.files.savedResourceDescription', { destination: normalizedDestination });
                } else if (normalizedDestination) {
                    description = t('flow.files.savedResourcesDescription', {
                        count,
                        destination: normalizedDestination,
                    });
                } else {
                    description = t('flow.files.savedResourcesRootDescription', { count });
                }

                toast.success(t('flow.files.savedToResources'), { description });

                return { kind: 'ok' };
            } catch (error) {
                if (!force && getApiErrorStatusCode(error) === 409) {
                    return { kind: 'conflict' };
                }

                const description = getApiErrorMessage(error, t('flow.files.saveAsResourceFailed'));

                toast.error(t('flow.files.saveAsResourceFailed'), { description });

                return { kind: 'error' };
            } finally {
                setIsPromoting(false);
            }
        },
        [flowId, t],
    );

    return {
        isPromoting,
        promote,
    };
}
