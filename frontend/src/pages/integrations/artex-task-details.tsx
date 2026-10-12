import { Download, RefreshCw } from 'lucide-react';
import { useCallback, useState } from 'react';

import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { useLocale } from '@/hooks/use-locale';

import { artexApi } from './artex-api';
import { artexErrorText, useArtexRequest } from './use-artex-request';

interface TaskProps {
    revision: number;
    taskID: string;
}

function Activity({ revision, taskID }: TaskProps) {
    const { t } = useLocale();
    const [cursors, setCursors] = useState([0]);
    const since = cursors[cursors.length - 1] ?? 0;
    const request = useCallback(
        (signal: AbortSignal) => {
            void revision;

            return artexApi.activity(taskID, since, signal);
        },
        [taskID, revision, since],
    );
    const { data, error, loading } = useArtexRequest(request);

    return (
        <div>
            <p className="text-muted-foreground mb-4 text-xs">{t('integrations.artex.activityHint')}</p>
            <RequestState
                error={error}
                loading={loading}
            />
            {data ? (
                <ol className="divide-y">
                    {(data.items ?? []).map((item) => (
                        <li
                            className="py-3 text-sm"
                            key={item.seq}
                        >
                            <div className="text-muted-foreground flex flex-wrap gap-3 text-xs">
                                <span className="font-mono">
                                    #{item.seq} · {item.worker}
                                </span>
                                <time>{item.ts}</time>
                            </div>
                            <p className="mt-1 break-words whitespace-pre-wrap">{item.summary}</p>
                        </li>
                    ))}
                </ol>
            ) : null}
            {data && !data.items?.length ? (
                <p className="text-muted-foreground">{t('integrations.artex.emptyActivity')}</p>
            ) : null}
            <div className="mt-4 flex justify-end gap-2">
                <Button
                    disabled={loading || cursors.length === 1}
                    onClick={() => setCursors((previous) => previous.slice(0, -1))}
                    size="sm"
                    variant="outline"
                >
                    {t('integrations.artex.previous')}
                </Button>
                <Button
                    disabled={loading || !data || data.cursor <= since || (data.items?.length ?? 0) < 100}
                    onClick={() => {
                        if (data) {
                            setCursors((previous) => [...previous, data.cursor]);
                        }
                    }}
                    size="sm"
                    variant="outline"
                >
                    {t('integrations.artex.next')}
                </Button>
            </div>
        </div>
    );
}

function ArtexTaskDetails(props: TaskProps & { canControl: boolean }) {
    const { t } = useLocale();

    return (
        <Tabs
            className="p-4"
            defaultValue="findings"
        >
            <TabsList aria-label={t('integrations.artex.selectedTask')}>
                <TabsTrigger value="findings">{t('integrations.artex.findings')}</TabsTrigger>
                <TabsTrigger value="assets">{t('integrations.artex.assets')}</TabsTrigger>
                <TabsTrigger value="activity">{t('integrations.artex.activity')}</TabsTrigger>
            </TabsList>
            <TabsContent value="findings">
                <Findings {...props} />
            </TabsContent>
            <TabsContent value="assets">
                <Assets {...props} />
            </TabsContent>
            <TabsContent value="activity">
                <Activity {...props} />
            </TabsContent>
        </Tabs>
    );
}

function Assets({ revision, taskID }: TaskProps) {
    const { t } = useLocale();
    const [page, setPage] = useState(1);
    const request = useCallback(
        (signal: AbortSignal) => {
            void revision;

            return artexApi.assets(taskID, page, signal);
        },
        [taskID, revision, page],
    );
    const { data, error, loading } = useArtexRequest(request);

    return (
        <div>
            <RequestState
                error={error}
                loading={loading}
            />
            {data ? (
                <>
                    <div className="divide-y">
                        {(data.assets ?? []).map((asset) => (
                            <div
                                className="flex flex-wrap justify-between gap-2 py-3 text-sm"
                                key={asset.id}
                            >
                                <span className="min-w-0 font-mono break-all">
                                    {asset.url ||
                                        asset.domain ||
                                        asset.app_name ||
                                        asset.ip ||
                                        asset.service_name ||
                                        '#' + asset.id}
                                </span>
                                <span className="text-muted-foreground">
                                    {asset.type}
                                    {asset.port ? ' · ' + asset.port : ''}
                                </span>
                            </div>
                        ))}
                    </div>
                    {!data.assets?.length ? (
                        <p className="text-muted-foreground">{t('integrations.artex.emptyAssets')}</p>
                    ) : null}
                    <Pager
                        page={page}
                        setPage={setPage}
                        total={data.total}
                    />
                </>
            ) : null}
        </div>
    );
}

function FindingDetails({ canControl, findingID }: { canControl: boolean; findingID: string }) {
    const { t } = useLocale();
    const [revision, setRevision] = useState(0);
    const [notes, setNotes] = useState('');
    const [busy, setBusy] = useState(false);
    const [mutationError, setMutationError] = useState('');
    const request = useCallback(
        async (signal: AbortSignal) => {
            void revision;
            const [finding, retests] = await Promise.all([
                artexApi.finding(findingID, signal),
                artexApi.retests(findingID, signal),
            ]);

            return { finding, retests: retests.retests ?? [] };
        },
        [findingID, revision],
    );
    const { data, error, loading } = useArtexRequest(request);
    const running = data?.retests.some((item) => ['pending', 'running'].includes(item.status));

    const startRetest = async () => {
        if (!canControl || busy || loading || running) {
            return;
        }

        setBusy(true);
        setMutationError('');

        try {
            await artexApi.retest(findingID, notes.trim());
            setNotes('');
            setRevision((value) => value + 1);
        } catch (caught) {
            setMutationError(artexErrorText(caught, t));
        } finally {
            setBusy(false);
        }
    };

    const download = () => {
        if (!data) {
            return;
        }

        const blob = new Blob(
            [JSON.stringify({ capturedAt: new Date().toISOString(), engine: 'artex', ...data }, null, 2)],
            { type: 'application/json' },
        );
        const url = URL.createObjectURL(blob);
        const anchor = document.createElement('a');
        anchor.href = url;
        anchor.download = 'artex-finding-' + findingID + '.json';
        anchor.click();
        window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    };

    return (
        <section
            aria-label={t('integrations.artex.evidence')}
            className="mt-4 rounded-md border p-4"
        >
            <div className="mb-4 flex flex-wrap justify-between gap-3">
                <h3 className="font-semibold">
                    {t('integrations.artex.evidence')} #{findingID}
                </h3>
                <div className="flex gap-2">
                    <Button
                        disabled={loading || busy}
                        onClick={() => setRevision((value) => value + 1)}
                        size="sm"
                        variant="outline"
                    >
                        <RefreshCw className="size-4" />
                        {t('integrations.artex.refreshRetests')}
                    </Button>
                    <Button
                        disabled={!data}
                        onClick={download}
                        size="sm"
                        variant="outline"
                    >
                        <Download className="size-4" />
                        {t('integrations.artex.exportEvidence')}
                    </Button>
                </div>
            </div>
            <RequestState
                error={error}
                loading={loading}
            />
            {data ? (
                <div className="flex flex-col gap-4">
                    <p className="text-sm break-words whitespace-pre-wrap">{data.finding.summary}</p>
                    <pre className="bg-muted max-h-80 overflow-auto rounded p-3 text-xs break-words whitespace-pre-wrap">
                        {data.finding.evidence || t('integrations.artex.noEvidence')}
                    </pre>
                    {data.finding.report ? (
                        <details>
                            <summary className="cursor-pointer text-sm">{t('integrations.artex.report')}</summary>
                            <pre className="mt-2 max-h-96 overflow-auto text-xs break-words whitespace-pre-wrap">
                                {data.finding.report}
                            </pre>
                        </details>
                    ) : null}
                    <h4 className="text-sm font-semibold">{t('integrations.artex.retests')}</h4>
                    {data.retests.map((item) => (
                        <div
                            className="border-l-2 pl-3 text-sm"
                            key={item.id}
                        >
                            <p className="font-medium">
                                #{item.id} · {t('integrations.artex.state.' + item.status)}
                                {item.verdict ? ' · ' + t('integrations.artex.verdict.' + item.verdict) : ''}
                            </p>
                            <p className="mt-1 break-words whitespace-pre-wrap">{item.summary}</p>
                            {item.evidence ? (
                                <pre className="mt-2 max-h-64 overflow-auto text-xs break-words whitespace-pre-wrap">
                                    {item.evidence}
                                </pre>
                            ) : null}
                        </div>
                    ))}
                    {!data.retests.length ? (
                        <p className="text-muted-foreground text-sm">{t('integrations.artex.noRetests')}</p>
                    ) : null}
                    {canControl ? (
                        <form
                            className="grid gap-3"
                            onSubmit={(event) => {
                                event.preventDefault();
                                void startRetest();
                            }}
                        >
                            <Label htmlFor="artex-retest-notes">{t('integrations.artex.retestNotes')}</Label>
                            <Textarea
                                disabled={busy || running}
                                id="artex-retest-notes"
                                maxLength={4000}
                                onChange={(event) => setNotes(event.target.value)}
                                value={notes}
                            />
                            <p className="text-muted-foreground text-xs">{t('integrations.artex.retestHint')}</p>
                            <Button
                                disabled={busy || running}
                                type="submit"
                            >
                                {running ? t('integrations.artex.retesting') : t('integrations.artex.startRetest')}
                            </Button>
                        </form>
                    ) : null}
                </div>
            ) : null}
            {mutationError ? (
                <p
                    className="text-destructive mt-3 text-sm"
                    role="alert"
                >
                    {mutationError}
                </p>
            ) : null}
        </section>
    );
}

function Findings({ canControl, revision, taskID }: TaskProps & { canControl: boolean }) {
    const { t } = useLocale();
    const [page, setPage] = useState(1);
    const [findingID, setFindingID] = useState('');
    const request = useCallback(
        (signal: AbortSignal) => {
            void revision;

            return artexApi.findings(taskID, page, signal);
        },
        [taskID, revision, page],
    );
    const { data, error, loading } = useArtexRequest(request);

    return (
        <div>
            <RequestState
                error={error}
                loading={loading}
            />
            {data ? (
                <>
                    <div className="divide-y">
                        {(data.items ?? []).map((item) => (
                            <button
                                aria-pressed={findingID === item.id}
                                className="hover:bg-muted/50 focus-visible:outline-primary block w-full rounded p-3 text-left focus-visible:outline-2"
                                key={item.id}
                                onClick={() => setFindingID(item.id)}
                                type="button"
                            >
                                <span className="flex flex-wrap justify-between gap-2 text-sm">
                                    <strong>
                                        {item.name || item.vulnclass || t('integrations.artex.unnamedFinding')}
                                    </strong>
                                    <span className="text-muted-foreground">
                                        {item.severity} · {item.status}
                                    </span>
                                </span>
                                <span className="text-muted-foreground mt-1 line-clamp-2 block text-xs">
                                    {item.summary}
                                </span>
                            </button>
                        ))}
                    </div>
                    {!data.items?.length ? (
                        <p className="text-muted-foreground">{t('integrations.artex.emptyFindings')}</p>
                    ) : null}
                    <Pager
                        page={page}
                        setPage={(next) => {
                            setFindingID('');
                            setPage(next);
                        }}
                        total={data.total}
                    />
                </>
            ) : null}
            {findingID ? (
                <FindingDetails
                    canControl={canControl}
                    findingID={findingID}
                    key={findingID}
                />
            ) : null}
        </div>
    );
}

function Pager({ page, setPage, total }: { page: number; setPage: (page: number) => void; total: number }) {
    const { t } = useLocale();

    return (
        <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm">
            <span className="text-muted-foreground">{t('integrations.artex.pageInfo', { page, total })}</span>
            <div className="flex gap-2">
                <Button
                    disabled={page === 1}
                    onClick={() => setPage(page - 1)}
                    size="sm"
                    variant="outline"
                >
                    {t('integrations.artex.previous')}
                </Button>
                <Button
                    disabled={page * 50 >= total}
                    onClick={() => setPage(page + 1)}
                    size="sm"
                    variant="outline"
                >
                    {t('integrations.artex.next')}
                </Button>
            </div>
        </div>
    );
}

function RequestState({ error, loading }: { error?: unknown; loading: boolean }) {
    const { t } = useLocale();

    return loading ? (
        <p role="status">{t('integrations.artex.loading')}</p>
    ) : error ? (
        <p
            className="text-destructive"
            role="alert"
        >
            {artexErrorText(error, t)}
        </p>
    ) : null;
}

export default ArtexTaskDetails;
