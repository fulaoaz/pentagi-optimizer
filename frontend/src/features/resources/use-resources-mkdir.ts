import { useCallback, useState } from 'react';
import { toast } from 'sonner';
import { z } from 'zod';

import type { Translate } from '@/lib/i18n';

import { useLocale } from '@/hooks/use-locale';
import { api, getApiErrorMessage } from '@/lib/axios';

import { RESOURCES_MKDIR_API_PATH } from './resources-constants';

export const createResourcesMkdirFormSchema = (t: Translate) =>
    z.object({
        path: z
            .string()
            .trim()
            .min(1, { message: t('resources.pathRequired') })
            .refine((value) => !value.startsWith('/'), { message: t('resources.pathRelative') })
            .refine((value) => !value.split('/').includes('..'), {
                message: t('resources.pathNoParentSegment'),
            }),
    });

export type ResourcesMkdirFormValues = z.infer<ReturnType<typeof createResourcesMkdirFormSchema>>;

interface MkdirRequestBody {
    path: string;
}

interface UseResourcesMkdirResult {
    isCreating: boolean;
    mkdir: (values: ResourcesMkdirFormValues) => Promise<boolean>;
}

/** Wraps `POST /resources/mkdir` (idempotent — returns existing dir on hit). */
export function useResourcesMkdir(): UseResourcesMkdirResult {
    const { t } = useLocale();
    const [isCreating, setIsCreating] = useState(false);

    const mkdir = useCallback(
        async ({ path }: ResourcesMkdirFormValues): Promise<boolean> => {
            setIsCreating(true);

            try {
                await api.post<void, MkdirRequestBody>(RESOURCES_MKDIR_API_PATH, { path: path.trim() });

                toast.success(t('resources.directoryCreated'), {
                    description: t('resources.createdAt', { path: `/${path.trim()}` }),
                });

                return true;
            } catch (error) {
                const description = getApiErrorMessage(error, t('resources.createDirectoryFailed'));

                toast.error(t('resources.createDirectoryFailed'), { description });

                return false;
            } finally {
                setIsCreating(false);
            }
        },
        [t],
    );

    return {
        isCreating,
        mkdir,
    };
}
