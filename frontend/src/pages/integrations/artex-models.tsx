import { useCallback, useState } from 'react';
import { Link } from 'react-router-dom';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useLocale } from '@/hooks/use-locale';
import { routes } from '@/lib/routes';

import type { ArtexBinding } from './artex-api';

import { artexApi } from './artex-api';
import { artexErrorText, useArtexRequest } from './use-artex-request';

const selectClass = 'border-input bg-background h-9 w-full min-w-0 rounded-md border px-3 text-sm';

function ArtexModels({ onSynced }: { onSynced: (binding: ArtexBinding) => void }) {
    const { t } = useLocale();
    const [revision, setRevision] = useState(0);
    const [providerRef, setProviderRef] = useState('');
    const [model, setModel] = useState('');
    const [maxTokens, setMaxTokens] = useState(0);
    const [effort, setEffort] = useState('');
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState('');
    const request = useCallback(
        async (signal: AbortSignal) => {
            void revision;
            const [providers, bindings] = await Promise.all([
                artexApi.modelProviders(signal),
                artexApi.modelBindings(signal),
            ]);

            return { bindings, providers };
        },
        [revision],
    );
    const { data, error: loadError, loading } = useArtexRequest(request);
    const source = data?.providers.find((item) => item.ref === providerRef);

    const sync = async () => {
        if (!source?.compatible || !model.trim() || busy) {
            return;
        }

        setBusy(true);
        setError('');

        try {
            const binding = await artexApi.syncModel({
                maxTokens,
                model: model.trim(),
                providerRef,
                reasoningEffort: effort,
            });
            onSynced(binding);
        } catch (caught) {
            setError(artexErrorText(caught, t));
        } finally {
            setBusy(false);
        }
    };

    const rotate = async (id: string) => {
        setBusy(true);
        setError('');

        try {
            const binding = await artexApi.rotateModelKey(id);
            setRevision((value) => value + 1);
            onSynced(binding);
        } catch (caught) {
            setError(artexErrorText(caught, t));
        } finally {
            setBusy(false);
        }
    };

    return (
        <div className="flex min-w-0 flex-col gap-4">
            <p className="text-muted-foreground text-sm">{t('integrations.artex.models.description')}</p>
            <Link
                className="text-sm underline"
                to={routes.settings.providers}
            >
                {t('integrations.artex.models.openProviders')}
            </Link>
            {loading ? <p role="status">{t('integrations.artex.loading')}</p> : null}
            {error || loadError ? (
                <p
                    className="text-destructive text-sm"
                    role="alert"
                >
                    {error || artexErrorText(loadError, t)}
                </p>
            ) : null}
            {data ? (
                <>
                    <div className="grid gap-2">
                        <Label htmlFor="artex-provider">{t('integrations.artex.models.provider')}</Label>
                        <select
                            className={selectClass}
                            disabled={busy}
                            id="artex-provider"
                            onChange={(event) => {
                                setProviderRef(event.target.value);
                                setModel(
                                    data.providers.find((item) => item.ref === event.target.value)?.defaultModel ?? '',
                                );
                            }}
                            value={providerRef}
                        >
                            <option value="">{t('integrations.artex.models.chooseProvider')}</option>
                            {data.providers.map((item) => (
                                <option
                                    disabled={!item.compatible}
                                    key={item.ref}
                                    value={item.ref}
                                >
                                    {item.name}
                                    {item.compatible
                                        ? ''
                                        : ' · ' + t('integrations.artex.models.reason.' + item.reason)}
                                </option>
                            ))}
                        </select>
                    </div>
                    <div className="grid gap-2">
                        <Label htmlFor="artex-model">{t('integrations.artex.models.model')}</Label>
                        <Input
                            disabled={busy || !source?.compatible}
                            id="artex-model"
                            list="artex-model-list"
                            maxLength={512}
                            onChange={(event) => setModel(event.target.value)}
                            value={model}
                        />
                        <datalist id="artex-model-list">
                            {source?.models.map((name) => (
                                <option
                                    key={name}
                                    value={name}
                                />
                            ))}
                        </datalist>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-2">
                        <div className="grid gap-2">
                            <Label htmlFor="artex-max-tokens">{t('integrations.artex.models.maxTokens')}</Label>
                            <Input
                                disabled={busy}
                                id="artex-max-tokens"
                                max={1000000}
                                min={0}
                                onChange={(event) => setMaxTokens(Number(event.target.value))}
                                type="number"
                                value={maxTokens}
                            />
                        </div>
                        <div className="grid gap-2">
                            <Label htmlFor="artex-effort">{t('integrations.artex.models.effort')}</Label>
                            <select
                                className={selectClass}
                                disabled={busy}
                                id="artex-effort"
                                onChange={(event) => setEffort(event.target.value)}
                                value={effort}
                            >
                                <option value="">{t('integrations.artex.models.engineDefault')}</option>
                                {['low', 'medium', 'high', 'xhigh', 'max'].map((value) => (
                                    <option
                                        key={value}
                                        value={value}
                                    >
                                        {value}
                                    </option>
                                ))}
                            </select>
                        </div>
                    </div>
                    <Button
                        disabled={busy || !source?.compatible || !model.trim()}
                        onClick={() => void sync()}
                        type="button"
                    >
                        {t('integrations.artex.models.sync')}
                    </Button>
                    <p className="text-muted-foreground text-xs">{t('integrations.artex.models.parametersHint')}</p>
                    <h3 className="mt-2 font-semibold">{t('integrations.artex.models.existing')}</h3>
                    <p className="text-muted-foreground text-xs">{t('integrations.artex.models.rotationHint')}</p>
                    <div className="max-h-56 divide-y overflow-auto">
                        {data.bindings.map((binding) => (
                            <div
                                className="flex flex-wrap items-center justify-between gap-3 py-3 text-sm"
                                key={binding.id}
                            >
                                <div className="min-w-0">
                                    <div className="font-medium break-all">{binding.spec.model}</div>
                                    <p className="text-muted-foreground text-xs">
                                        {binding.revision.slice(0, 8)} ·{' '}
                                        {t('integrations.artex.models.state.' + binding.state)}
                                    </p>
                                </div>
                                <Button
                                    disabled={busy || !binding.remoteProfileId || binding.state === 'revoked'}
                                    onClick={() => void rotate(binding.id)}
                                    size="sm"
                                    type="button"
                                    variant="outline"
                                >
                                    {t('integrations.artex.models.rotate')}
                                </Button>
                            </div>
                        ))}
                    </div>
                </>
            ) : null}
        </div>
    );
}

export default ArtexModels;
