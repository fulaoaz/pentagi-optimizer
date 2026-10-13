import type { FormEvent } from 'react';

import { Pause, Play, Plus, Radar, RefreshCw } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useLocale } from '@/hooks/use-locale';
import { routes } from '@/lib/routes';
import { useUser } from '@/providers/user-provider';

import { artexApi } from './artex-api';
import ArtexModels from './artex-models';
import ArtexTaskDetails from './artex-task-details';
import { artexErrorText, useArtexRequest } from './use-artex-request';

function ArtexWorkbench() {
    const { authInfo } = useUser();
    const { t } = useLocale();

    return authInfo?.privileges?.includes('flows.admin') ? (
        <ConnectedWorkbench />
    ) : (
        <div className="p-4">
            <Alert>
                <AlertTitle>{t('integrations.artex.title')}</AlertTitle>
                <AlertDescription>{t('integrations.artex.adminOnly')}</AlertDescription>
            </Alert>
        </div>
    );
}

function ConnectedWorkbench() {
    const { t } = useLocale();
    const [searchParams, setSearchParams] = useSearchParams();
    const [revision, setRevision] = useState(0);
    const [selectedID, setSelectedID] = useState(() => searchParams.get('task') ?? '');
    const [creating, setCreating] = useState(false);
    const [modeling, setModeling] = useState(false);
    const [returnToCreate, setReturnToCreate] = useState(false);
    const [bindingID, setBindingID] = useState('');
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState('');
    const [description, setDescription] = useState('');
    const [goal, setGoal] = useState('');
    const [timeout, setTimeoutValue] = useState(1800);
    const request = useCallback(
        async (signal: AbortSignal) => {
            // Revision only changes when the user requests a fresh snapshot.
            void revision;
            const status = await artexApi.status(signal);
            const tasks = status.healthy ? ((await artexApi.tasks(signal)).tasks ?? []) : [];
            const bindings = status.healthy && status.modelSharingAvailable ? await artexApi.modelBindings(signal) : [];

            return { bindings, status, tasks };
        },
        [revision],
    );
    const { data, error: loadError, loading } = useArtexRequest(request, true);
    const tasks = data?.tasks ?? [];
    const bindings = data?.bindings ?? [];
    const status = data?.status;
    const selected = tasks.find((task) => task.id === selectedID) ?? tasks[0];
    const connectionError = status?.errorCode ? t('integrations.artex.error.' + status.errorCode) : '';
    const message = error || (loadError ? artexErrorText(loadError, t) : connectionError);

    const refresh = () => {
        setError('');
        setRevision((value) => value + 1);
    };

    useEffect(() => {
        const requested = searchParams.get('task');
        if (requested && tasks.some((task) => task.id === requested)) {
            setSelectedID(requested);
        }
    }, [searchParams, tasks]);

    const createTask = async (event: FormEvent<HTMLFormElement>) => {
        event.preventDefault();

        if (!status?.canCreate || busy || loading || !bindingID) {
            return;
        }

        setBusy(true);
        setError('');

        try {
            const task = await artexApi.create(description.trim(), goal.trim(), timeout, bindingID);
            setSelectedID(task.id);
            setSearchParams({ task: task.id }, { replace: true });
            setCreating(false);
            setDescription('');
            setGoal('');
            refresh();
        } catch (caught) {
            setError(artexErrorText(caught, t));
        } finally {
            setBusy(false);
        }
    };

    const control = async (action: 'pause' | 'resume') => {
        if (!selected || !status?.canControl || busy || loading) {
            return;
        }

        setBusy(true);
        setError('');

        try {
            await artexApi.control(selected.id, action);
            refresh();
        } catch (caught) {
            setError(artexErrorText(caught, t));
        } finally {
            setBusy(false);
        }
    };

    return (
        <div className="flex min-w-0 flex-col gap-5 p-4">
            <header className="flex flex-wrap items-start justify-between gap-4">
                <div>
                    <h1 className="flex items-center gap-2 text-2xl font-semibold">
                        <Radar className="size-6" />
                        {t('integrations.artex.title')}
                    </h1>
                    <p className="text-muted-foreground mt-2 text-sm">{t('integrations.artex.description')}</p>
                    <div className="mt-3 flex items-center gap-3 text-xs">
                        <Badge variant={status?.healthy ? 'default' : 'secondary'}>
                            {loading
                                ? t('integrations.artex.loading')
                                : status?.healthy
                                  ? t('integrations.artex.online')
                                  : status?.enabled
                                    ? t('integrations.artex.unreachable')
                                    : t('integrations.artex.notConfigured')}
                        </Badge>
                        {status?.version ? (
                            <span className="text-muted-foreground font-mono">{status.version}</span>
                        ) : null}
                        <Link
                            className="underline underline-offset-4"
                            to={routes.flows}
                        >
                            {t('integrations.artex.nativeTasks')}
                        </Link>
                    </div>
                </div>
                <div className="flex flex-wrap gap-2">
                    {status?.canShareModels ? (
                        <Button
                            disabled={busy || loading}
                            onClick={() => {
                                setReturnToCreate(false);
                                setModeling(true);
                            }}
                            variant="outline"
                        >
                            {t('integrations.artex.models.manage')}
                        </Button>
                    ) : null}
                    <Button
                        disabled={busy || loading}
                        onClick={refresh}
                        variant="outline"
                    >
                        <RefreshCw className="size-4" />
                        {t('integrations.artex.refresh')}
                    </Button>
                    <Button
                        disabled={busy || !status?.canCreate}
                        onClick={() => setCreating(true)}
                    >
                        <Plus className="size-4" />
                        {t('integrations.artex.createTask')}
                    </Button>
                </div>
            </header>
            {message ? (
                <Alert variant="destructive">
                    <AlertTitle>{t('integrations.artex.errorTitle')}</AlertTitle>
                    <AlertDescription>{message}</AlertDescription>
                </Alert>
            ) : null}
            {loading ? <p role="status">{t('integrations.artex.loading')}</p> : null}
            {!loading && !status?.enabled && !loadError ? (
                <Alert>
                    <AlertTitle>{t('integrations.artex.notConfigured')}</AlertTitle>
                    <AlertDescription>
                        <p>{t('integrations.artex.setupHint')}</p>
                        <Link
                            className="underline"
                            to={routes.settings.runtimeConfig}
                        >
                            {t('integrations.artex.openSettings')}
                        </Link>
                    </AlertDescription>
                </Alert>
            ) : null}
            {status?.healthy && !status.canCreate && !status.canControl ? (
                <p className="text-muted-foreground text-sm">{t('integrations.artex.readOnly')}</p>
            ) : null}
            {status?.healthy ? (
                <div className="grid min-w-0 gap-5 xl:grid-cols-[18rem_minmax(0,1fr)]">
                    <section
                        aria-label={t('integrations.artex.taskList')}
                        className="rounded-lg border"
                    >
                        <h2 className="border-b p-4 font-semibold">
                            {t('integrations.artex.taskList')}{' '}
                            <span className="text-muted-foreground">({tasks.length})</span>
                        </h2>
                        <div className="max-h-[65vh] overflow-y-auto p-2">
                            {tasks.map((task) => (
                                <button
                                    aria-pressed={selected?.id === task.id}
                                    className={
                                        'focus-visible:outline-primary mb-1 w-full rounded-md p-3 text-left focus-visible:outline-2 ' +
                                        (selected?.id === task.id ? 'bg-muted' : 'hover:bg-muted/50')
                                    }
                                    disabled={busy}
                                    key={task.id}
                                    onClick={() => {
                                        setSelectedID(task.id);
                                        setSearchParams({ task: task.id }, { replace: true });
                                    }}
                                    type="button"
                                >
                                    <span className="block truncate font-medium">
                                        {task.name || task.description || t('integrations.artex.unnamedTask')}
                                    </span>
                                    <span className="text-muted-foreground mt-2 flex justify-between gap-2 text-xs">
                                        <span>#{task.id}</span>
                                        <span>{t('integrations.artex.state.' + task.status)}</span>
                                    </span>
                                </button>
                            ))}
                            {!tasks.length ? (
                                <p className="text-muted-foreground p-3 text-sm">
                                    {t('integrations.artex.emptyTasks')}
                                </p>
                            ) : null}
                        </div>
                    </section>
                    {selected ? (
                        <section className="min-w-0 rounded-lg border">
                            <div className="flex flex-wrap items-start justify-between gap-3 border-b p-4">
                                <div className="min-w-0 flex-1">
                                    <h2 className="font-semibold break-words">
                                        {selected.name || selected.description}
                                    </h2>
                                    <p className="text-muted-foreground mt-2 text-sm break-words whitespace-pre-wrap">
                                        {selected.goal}
                                    </p>
                                </div>
                                <div className="flex gap-2">
                                    <Button
                                        disabled={busy || !status.canControl || selected.status !== 'paused'}
                                        onClick={() => void control('resume')}
                                        size="sm"
                                        variant="outline"
                                    >
                                        <Play className="size-4" />
                                        {t('integrations.artex.resume')}
                                    </Button>
                                    <Button
                                        disabled={
                                            busy ||
                                            !status.canControl ||
                                            !['created', 'queued', 'running'].includes(selected.status)
                                        }
                                        onClick={() => void control('pause')}
                                        size="sm"
                                        variant="outline"
                                    >
                                        <Pause className="size-4" />
                                        {t('integrations.artex.pause')}
                                    </Button>
                                </div>
                            </div>
                            <ArtexTaskDetails
                                canControl={status.canControl && !loading}
                                key={selected.id}
                                revision={revision}
                                taskID={selected.id}
                            />
                        </section>
                    ) : null}
                </div>
            ) : null}
            <Dialog
                onOpenChange={(value) => {
                    if (!busy) {
                        setCreating(value);
                    }
                }}
                open={creating}
            >
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>{t('integrations.artex.createTask')}</DialogTitle>
                    </DialogHeader>
                    <form
                        className="flex flex-col gap-4"
                        onSubmit={createTask}
                    >
                        <p className="text-muted-foreground text-sm">{t('integrations.artex.createTaskDescription')}</p>
                        <div className="grid gap-2">
                            <Label htmlFor="artex-binding">{t('integrations.artex.models.taskBinding')}</Label>
                            <select
                                className="border-input bg-background h-9 w-full min-w-0 rounded-md border px-3 text-sm"
                                disabled={busy || loading}
                                id="artex-binding"
                                onChange={(event) => setBindingID(event.target.value)}
                                required
                                value={bindingID}
                            >
                                <option value="">{t('integrations.artex.models.chooseBinding')}</option>
                                {bindings
                                    .filter((item) => item.state === 'ready')
                                    .map((item) => (
                                        <option
                                            key={item.id}
                                            value={item.id}
                                        >
                                            {item.spec.model} · {item.revision.slice(0, 8)}
                                        </option>
                                    ))}
                            </select>
                            {status?.canShareModels ? (
                                <Button
                                    disabled={busy}
                                    onClick={() => {
                                        setReturnToCreate(true);
                                        setCreating(false);
                                        setModeling(true);
                                    }}
                                    type="button"
                                    variant="outline"
                                >
                                    {t('integrations.artex.models.manage')}
                                </Button>
                            ) : null}
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="artex-description">{t('integrations.artex.descriptionPlaceholder')}</Label>
                            <Input
                                id="artex-description"
                                maxLength={8000}
                                onChange={(event) => setDescription(event.target.value)}
                                required
                                value={description}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="artex-goal">{t('integrations.artex.goalLabel')}</Label>
                            <Textarea
                                id="artex-goal"
                                maxLength={8000}
                                onChange={(event) => setGoal(event.target.value)}
                                required
                                value={goal}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="artex-timeout">{t('integrations.artex.timeoutLabel')}</Label>
                            <Input
                                id="artex-timeout"
                                max={86400}
                                min={1}
                                onChange={(event) => setTimeoutValue(Number(event.target.value))}
                                required
                                type="number"
                                value={timeout}
                            />
                        </div>
                        {error ? (
                            <p
                                className="text-destructive text-sm"
                                role="alert"
                            >
                                {error}
                            </p>
                        ) : null}
                        <Button
                            disabled={
                                busy ||
                                loading ||
                                !status?.canCreate ||
                                !bindingID ||
                                !description.trim() ||
                                !goal.trim()
                            }
                            type="submit"
                        >
                            {busy ? t('integrations.artex.submitting') : t('integrations.artex.createTask')}
                        </Button>
                    </form>
                </DialogContent>
            </Dialog>
            <Dialog
                onOpenChange={(open) => {
                    setModeling(open);

                    if (!open && returnToCreate) {
                        setCreating(true);
                    }
                }}
                open={modeling}
            >
                <DialogContent className="max-h-[90vh] overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>{t('integrations.artex.models.manage')}</DialogTitle>
                    </DialogHeader>
                    {modeling ? (
                        <ArtexModels
                            onSynced={(binding) => {
                                setBindingID(binding.id);
                                setModeling(false);

                                if (returnToCreate) {
                                    setCreating(true);
                                }

                                refresh();
                            }}
                        />
                    ) : null}
                </DialogContent>
            </Dialog>
        </div>
    );
}

export default ArtexWorkbench;
