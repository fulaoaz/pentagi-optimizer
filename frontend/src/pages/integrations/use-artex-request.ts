import { useEffect, useState } from 'react';

import type { Translate } from '@/lib/i18n';

import { getApiErrorMessage } from '@/lib/axios';

export function artexErrorText(error: unknown, t: Translate): string {
    if (error instanceof Error && error.message === 'Artex.Response') {
        return t('integrations.artex.error.Artex.Response');
    }

    return getApiErrorMessage(error, t('integrations.artex.loadError'));
}

export function useArtexRequest<T>(request: (signal: AbortSignal) => Promise<T>, keepPrevious = false) {
    const [result, setResult] = useState<{
        data?: T;
        error?: unknown;
        request: typeof request;
    }>();

    useEffect(() => {
        const controller = new AbortController();

        request(controller.signal).then(
            (data) => {
                if (!controller.signal.aborted) {
                    setResult({ data, request });
                }
            },
            (error: unknown) => {
                if (!controller.signal.aborted) {
                    setResult({ error, request });
                }
            },
        );

        return () => controller.abort();
    }, [request]);

    return result?.request === request
        ? { data: result.data, error: result.error, loading: false }
        : { data: keepPrevious ? result?.data : undefined, error: undefined, loading: true };
}
