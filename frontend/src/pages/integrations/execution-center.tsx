import { ListChecks, RefreshCw, SearchCheck } from 'lucide-react';
import { useCallback, useState } from 'react';
import { Link } from 'react-router-dom';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge, type BadgeVariant } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useLocale } from '@/hooks/use-locale';
import { routes } from '@/lib/routes';
import { useUser } from '@/providers/user-provider';

import { type UnifiedRun, executionApi } from './artex-api';
import { artexErrorText, useArtexRequest } from './use-artex-request';

function statusVariant(status: string): BadgeVariant {
    if (status === 'completed' || status === 'finished') return 'green';
    if (status === 'failed' || status === 'timeout') return 'red';
    if (status === 'running') return 'blue';
    if (status === 'sync_unknown' || status === 'unknown') return 'orange';
    return 'secondary';
}

function ExecutionCenter() {
    const { t } = useLocale();
    const { authInfo } = useUser();
    const [revision, setRevision] = useState(0);
    const [busyID, setBusyID] = useState('');
    const request = useCallback(
        (signal: AbortSignal) => {
            void revision;
            return executionApi.list(signal);
        },
        [revision],
    );
    const { data, error, loading } = useArtexRequest(request, true);
    const items = data?.items ?? [];

    const refresh = () => setRevision((value) => value + 1);
    const reconcile = async (run: UnifiedRun) => {
        const localID = run.id.startsWith('artex:') ? run.id.slice('artex:'.length) : '';
        if (!localID || busyID) return;
        setBusyID(run.id);
        try {
            await executionApi.reconcile(localID);
            refresh();
        } finally {
            setBusyID('');
        }
    };

    return (
        <div className="flex min-w-0 flex-col gap-5 p-4">
            <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                    <div className="flex items-center gap-2">
                        <ListChecks className="size-5" />
                        <h1 className="text-2xl font-semibold">{t('executionCenter.title')}</h1>
                    </div>
                    <p className="text-muted-foreground mt-1 text-sm">{t('executionCenter.description')}</p>
                </div>
                <div className="flex flex-wrap gap-2">
                    <Button disabled={loading} onClick={refresh} size="sm" variant="outline">
                        <RefreshCw className="mr-2 size-4" />
                        {t('executionCenter.refresh')}
                    </Button>
                    <Button asChild size="sm">
                        <Link to={routes.newFlow}>{t('executionCenter.newPentagi')}</Link>
                    </Button>
                    {authInfo?.privileges?.includes('flows.admin') ? (
                        <Button asChild size="sm" variant="secondary">
                            <Link to={routes.artex}>{t('executionCenter.newArtex')}</Link>
                        </Button>
                    ) : null}
                </div>
            </div>

            {error ? (
                <Alert variant="destructive">
                    <AlertTitle>{t('executionCenter.loadError')}</AlertTitle>
                    <AlertDescription>{artexErrorText(error, t)}</AlertDescription>
                </Alert>
            ) : null}

            <Card>
                <CardHeader>
                    <CardTitle>{t('executionCenter.runs')}</CardTitle>
                    <CardDescription>{t('executionCenter.runsHint')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {loading && !items.length ? (
                        <p className="text-muted-foreground text-sm">{t('executionCenter.loading')}</p>
                    ) : !items.length ? (
                        <div className="text-muted-foreground flex flex-col items-center gap-2 py-12 text-sm">
                            <SearchCheck className="size-8" />
                            <span>{t('executionCenter.empty')}</span>
                        </div>
                    ) : (
                        <div className="divide-y rounded-md border">
                            {items.map((run) => {
                                const needsReconcile = run.engine === 'artex' && run.submitState === 'uncertain';
                                return (
                                    <div className="flex flex-wrap items-center justify-between gap-3 p-4" key={run.id}>
                                        <div className="min-w-0 flex-1">
                                            <div className="flex flex-wrap items-center gap-2">
                                                <Badge variant={run.engine === 'artex' ? 'purple' : 'blue'}>
                                                    {run.engine === 'artex' ? 'ARTEX' : 'PentAGI'}
                                                </Badge>
                                                <Link className="truncate font-medium hover:underline" to={run.detailPath}>
                                                    {run.title || t('executionCenter.untitled')}
                                                </Link>
                                                <Badge variant={statusVariant(run.status)}>
                                                    {t('executionCenter.status.' + run.status, { fallback: run.status })}
                                                </Badge>
                                            </div>
                                            <div className="text-muted-foreground mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs">
                                                <span>{t('executionCenter.rawStatus')}: {run.rawStatus || run.status}</span>
                                                {run.model ? <span>{t('executionCenter.model')}: {run.model}</span> : null}
                                                {run.engineTaskId ? <span>ID: {run.engineTaskId}</span> : null}
                                                <span>{new Date(run.updatedAt || run.createdAt).toLocaleString()}</span>
                                            </div>
                                        </div>
                                        {needsReconcile ? (
                                            <Button disabled={busyID !== ''} onClick={() => void reconcile(run)} size="sm" variant="outline">
                                                <SearchCheck className="mr-2 size-4" />
                                                {busyID === run.id ? t('executionCenter.reconciling') : t('executionCenter.reconcile')}
                                            </Button>
                                        ) : null}
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

export default ExecutionCenter;
