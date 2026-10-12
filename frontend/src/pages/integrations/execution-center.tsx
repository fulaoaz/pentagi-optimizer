import { ListChecks, RefreshCw, SearchCheck } from 'lucide-react';
import type { FormEvent } from 'react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge, type BadgeVariant } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useLocale } from '@/hooks/use-locale';
import { useFlows } from '@/providers/flows-provider';
import { useProviders } from '@/providers/providers-provider';
import { routes } from '@/lib/routes';
import { useUser } from '@/providers/user-provider';

import { artexApi, type UnifiedRun, executionApi } from './artex-api';
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
    const navigate = useNavigate();
    const { createFlow } = useFlows();
    const { selectedProvider } = useProviders();
    const [revision, setRevision] = useState(0);
    const [busyID, setBusyID] = useState('');
    const [createOpen, setCreateOpen] = useState(false);
    const [engine, setEngine] = useState<'artex' | 'pentagi'>('pentagi');
    const [message, setMessage] = useState('');
    const [description, setDescription] = useState('');
    const [goal, setGoal] = useState('');
    const [bindingID, setBindingID] = useState('');
    const [createError, setCreateError] = useState('');
    const [creating, setCreating] = useState(false);
    const isAdmin = authInfo?.privileges?.includes('flows.admin') ?? false;
    const request = useCallback(
        (signal: AbortSignal) => {
            void revision;
            return executionApi.list(signal);
        },
        [revision],
    );
    const { data, error, loading } = useArtexRequest(request, true);
    const bindingRequest = useCallback((signal: AbortSignal) => artexApi.modelBindings(signal), []);
    const { data: bindings, loading: bindingsLoading } = useArtexRequest(
        bindingRequest,
        createOpen && engine === 'artex' && isAdmin,
    );
    const items = data?.items ?? [];

    useEffect(() => {
        if (!bindingID && bindings?.length) {
            setBindingID(bindings.find((item) => item.state === 'ready')?.id ?? bindings[0].id);
        }
    }, [bindingID, bindings]);

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

    const createExecution = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        if (creating) return;
        setCreating(true);
        setCreateError('');
        try {
            if (engine === 'pentagi') {
                const flowID = await createFlow({
                    message: message.trim(),
                    providerName: selectedProvider?.name ?? '',
                    resourceIds: [],
                    useAgents: false,
                });
                if (!flowID) throw new Error('Flow could not be created');
                setCreateOpen(false);
                navigate(routes.flow(flowID));
            } else {
                if (!bindingID) throw new Error(t('executionCenter.bindingRequired'));
                await artexApi.create(description.trim(), goal.trim() || message.trim(), 1800, bindingID);
                setCreateOpen(false);
                setMessage('');
                setDescription('');
                setGoal('');
                setRevision((value) => value + 1);
            }
        } catch (caught) {
            setCreateError(artexErrorText(caught, t));
        } finally {
            setCreating(false);
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
                    <Button onClick={() => setCreateOpen(true)} size="sm">
                        {t('executionCenter.newExecution')}
                    </Button>
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
            <Dialog
                onOpenChange={(open) => {
                    if (!creating) setCreateOpen(open);
                }}
                open={createOpen}
            >
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('executionCenter.newExecution')}</DialogTitle>
                    </DialogHeader>
                    <form className="grid gap-4" onSubmit={(event) => void createExecution(event)}>
                        <div className="grid gap-2">
                            <Label>{t('executionCenter.engine')}</Label>
                            <Select
                                onValueChange={(value: 'artex' | 'pentagi') => {
                                    setEngine(value);
                                    setCreateError('');
                                }}
                                value={engine}
                            >
                                <SelectTrigger><SelectValue /></SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="pentagi">PentAGI</SelectItem>
                                    {isAdmin ? <SelectItem value="artex">ARTEX</SelectItem> : null}
                                </SelectContent>
                            </Select>
                        </div>
                        {engine === 'pentagi' ? (
                            <>
                                <div className="grid gap-2">
                                    <Label htmlFor="execution-message">{t('executionCenter.goal')}</Label>
                                    <Textarea
                                        id="execution-message"
                                        minLength={1}
                                        onChange={(event) => setMessage(event.target.value)}
                                        placeholder={t('executionCenter.goalPlaceholder')}
                                        value={message}
                                    />
                                    <p className="text-muted-foreground text-xs">
                                        {selectedProvider?.name
                                            ? t('executionCenter.providerUsed', { provider: selectedProvider.name })
                                            : t('executionCenter.providerMissing')}
                                    </p>
                                </div>
                            </>
                        ) : (
                            <>
                                <div className="grid gap-2">
                                    <Label htmlFor="execution-description">{t('executionCenter.descriptionField')}</Label>
                                    <Input
                                        id="execution-description"
                                        onChange={(event) => setDescription(event.target.value)}
                                        value={description}
                                    />
                                </div>
                                <div className="grid gap-2">
                                    <Label htmlFor="execution-goal">{t('executionCenter.goal')}</Label>
                                    <Textarea
                                        id="execution-goal"
                                        minLength={1}
                                        onChange={(event) => setGoal(event.target.value)}
                                        value={goal}
                                    />
                                </div>
                                <div className="grid gap-2">
                                    <Label>{t('executionCenter.modelBinding')}</Label>
                                    <Select onValueChange={setBindingID} value={bindingID}>
                                        <SelectTrigger><SelectValue placeholder={t('executionCenter.selectBinding')} /></SelectTrigger>
                                        <SelectContent>
                                            {(bindings ?? []).map((binding) => (
                                                <SelectItem key={binding.id} value={binding.id}>
                                                    {binding.spec.model} · {binding.state}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                    {bindingsLoading ? <p className="text-muted-foreground text-xs">{t('executionCenter.loading')}</p> : null}
                                </div>
                            </>
                        )}
                        {createError ? <p className="text-destructive text-sm">{createError}</p> : null}
                        <Button disabled={creating || (engine === 'pentagi' ? !message.trim() || !selectedProvider : !bindingID || !goal.trim())} type="submit">
                            {creating ? t('executionCenter.creating') : t('executionCenter.create')}
                        </Button>
                    </form>
                </DialogContent>
            </Dialog>
        </div>
    );
}

export default ExecutionCenter;
