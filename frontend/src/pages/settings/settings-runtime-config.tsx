import { useMutation, useQuery } from '@apollo/client/react';
import { AlertCircle, ArrowRight, CheckCircle2, Info, Save, Search, Settings2, ShieldCheck } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { InputPassword } from '@/components/ui/input-password';
import { Label } from '@/components/ui/label';
import { Skeleton } from '@/components/ui/skeleton';
import { Textarea } from '@/components/ui/textarea';
import { RuntimeConfigDocument, UpdateRuntimeConfigDocument } from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';

export const runtimeConfigCategories = [
    { category: '核心运行时', slug: 'core', titleKey: 'settings.runtimeConfig.categories.core' },
    { category: '模型提供商与嵌入', slug: 'models', titleKey: 'settings.runtimeConfig.categories.models' },
    { category: 'MCP 安全桥', slug: 'mcp', titleKey: 'settings.runtimeConfig.categories.mcp' },
    { category: '容器与执行环境', slug: 'execution', titleKey: 'settings.runtimeConfig.categories.execution' },
    { category: '服务与网络', slug: 'network', titleKey: 'settings.runtimeConfig.categories.network' },
    { category: '搜索与安全情报', slug: 'search', titleKey: 'settings.runtimeConfig.categories.search' },
    { category: '代理行为', slug: 'agents', titleKey: 'settings.runtimeConfig.categories.agents' },
    { category: '观测与知识', slug: 'observability', titleKey: 'settings.runtimeConfig.categories.observability' },
    { category: '认证与安全', slug: 'security', titleKey: 'settings.runtimeConfig.categories.security' },
    { category: '数据库', slug: 'database', titleKey: 'settings.runtimeConfig.categories.database' },
] as const;

type RuntimeConfigEntry = {
    category: string;
    configured: boolean;
    defaultValue: string;
    description: string;
    key: string;
    kind: string;
    restartRequired: boolean;
    sensitive: boolean;
    value: string;
};

export default function SettingsRuntimeConfig() {
    const { t } = useLocale();
    const navigate = useNavigate();
    const { category: categorySlug } = useParams<{ category?: string }>();
    const [query, setQuery] = useState('');
    const [draftOverrides, setDraftOverrides] = useState<Record<string, string>>({});
    const [dirtyKeys, setDirtyKeys] = useState<Set<string>>(new Set());
    const { data, error, loading, refetch } = useQuery(RuntimeConfigDocument);
    const [updateRuntimeConfig, { error: updateError, loading: isSaving }] = useMutation(UpdateRuntimeConfigDocument);

    const entries = useMemo(() => (data?.runtimeConfig ?? []) as RuntimeConfigEntry[], [data?.runtimeConfig]);
    const activeCategory = runtimeConfigCategories.find((item) => item.slug === categorySlug);
    const normalizedQuery = query.trim().toLowerCase();
    const drafts = useMemo(
        () =>
            Object.fromEntries(
                entries.map((entry) => [
                    entry.key,
                    entry.key in draftOverrides ? draftOverrides[entry.key] : entry.sensitive ? '' : entry.value,
                ]),
            ),
        [draftOverrides, entries],
    );

    const visibleEntries = useMemo(
        () =>
            entries.filter((entry) => {
                if (activeCategory && entry.category !== activeCategory.category) {
                    return false;
                }

                if (!normalizedQuery) {
                    return true;
                }

                return [entry.key, entry.description, entry.value].join(' ').toLowerCase().includes(normalizedQuery);
            }),
        [activeCategory, entries, normalizedQuery],
    );

    const categoryCounts = useMemo(() => {
        const counts = new Map<string, number>();

        for (const entry of entries) {
            counts.set(entry.category, (counts.get(entry.category) ?? 0) + 1);
        }

        return counts;
    }, [entries]);

    const handleDraftChange = (key: string, value: string) => {
        setDraftOverrides((current) => ({ ...current, [key]: value }));
        setDirtyKeys((current) => new Set(current).add(key));
    };

    const handleSave = async () => {
        const updates = [...dirtyKeys].map((key) => ({ key, value: drafts[key] ?? '' }));

        if (updates.length === 0) {
            return;
        }

        try {
            await updateRuntimeConfig({ variables: { updates } });
            setDirtyKeys(new Set());
            setDraftOverrides({});
            await refetch();
        } catch {
            // Apollo exposes the error in updateError so the form stays editable.
        }
    };

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>{t('settings.runtimeConfig.loadError')}</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    const pageTitle = activeCategory ? t(activeCategory.titleKey) : t('settings.runtimeConfig.title');

    return (
        <div className="mx-auto flex w-full max-w-7xl flex-col gap-4">
            <Card>
                <CardHeader>
                    <div className="flex flex-wrap items-start justify-between gap-4">
                        <div>
                            <CardTitle className="flex items-center gap-2">
                                <Settings2 className="size-5" />
                                {pageTitle}
                            </CardTitle>
                            <CardDescription className="mt-2 max-w-3xl">
                                {t('settings.runtimeConfig.description')}
                            </CardDescription>
                        </div>
                        <Button
                            disabled={isSaving || dirtyKeys.size === 0}
                            onClick={handleSave}
                        >
                            <Save className="size-4" />
                            {isSaving ? t('common.saving') : t('common.saveChanges')}
                        </Button>
                    </div>
                </CardHeader>
                <CardContent className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="text-muted-foreground flex items-center gap-2 text-sm">
                        <Info className="size-4" />
                        {t('settings.runtimeConfig.summary', {
                            configured: entries.filter((entry) => entry.configured).length,
                            total: entries.length,
                        })}
                    </div>
                    <div className="relative w-full sm:max-w-xs">
                        <Search className="text-muted-foreground absolute top-2.5 left-2.5 size-4" />
                        <Input
                            className="pl-8"
                            onChange={(event) => setQuery(event.target.value)}
                            placeholder={t('settings.runtimeConfig.searchPlaceholder')}
                            value={query}
                        />
                    </div>
                </CardContent>
            </Card>

            {updateError && (
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>{t('settings.runtimeConfig.saveError')}</AlertTitle>
                    <AlertDescription>{updateError.message}</AlertDescription>
                </Alert>
            )}

            {!activeCategory && !loading && (
                <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                    {runtimeConfigCategories.map((item) => {
                        const count = categoryCounts.get(item.category) ?? 0;

                        return (
                            <Link
                                className="group"
                                key={item.slug}
                                to={`/settings/runtime-config/${item.slug}`}
                            >
                                <Card className="h-full transition-colors group-hover:border-primary/60">
                                    <CardHeader>
                                        <CardTitle className="flex items-center justify-between text-base">
                                            {t(item.titleKey)}
                                            <ArrowRight className="text-muted-foreground size-4 transition-transform group-hover:translate-x-1" />
                                        </CardTitle>
                                        <CardDescription>
                                            {t('settings.runtimeConfig.categorySummary', { count })}
                                        </CardDescription>
                                    </CardHeader>
                                </Card>
                            </Link>
                        );
                    })}
                </div>
            )}

            {activeCategory && (
                <div className="flex flex-wrap items-center gap-2">
                    <Button
                        onClick={() => navigate('/settings/runtime-config')}
                        size="sm"
                        variant="outline"
                    >
                        {t('settings.runtimeConfig.allCategories')}
                    </Button>
                    {runtimeConfigCategories.map((item) => (
                        <Button
                            asChild
                            key={item.slug}
                            size="sm"
                            variant={item.slug === activeCategory.slug ? 'secondary' : 'ghost'}
                        >
                            <Link to={`/settings/runtime-config/${item.slug}`}>{t(item.titleKey)}</Link>
                        </Button>
                    ))}
                </div>
            )}

            {loading && <Skeleton className="h-64 w-full" />}

            {!loading && activeCategory && visibleEntries.length === 0 && (
                <Card>
                    <CardContent className="text-muted-foreground py-10 text-center text-sm">
                        {t('settings.runtimeConfig.empty')}
                    </CardContent>
                </Card>
            )}

            {!loading && activeCategory && visibleEntries.length > 0 && (
                <div className="grid gap-4 md:grid-cols-2">
                    {visibleEntries.map((entry) => (
                        <RuntimeConfigField
                            dirty={dirtyKeys.has(entry.key)}
                            entry={entry}
                            key={entry.key}
                            onChange={handleDraftChange}
                            value={drafts[entry.key] ?? ''}
                        />
                    ))}
                </div>
            )}

            {activeCategory && (
                <Card className="border-amber-500/30 bg-amber-500/5">
                    <CardContent className="text-muted-foreground flex items-start gap-3 py-4 text-sm">
                        <ShieldCheck className="mt-0.5 size-4 shrink-0 text-amber-600" />
                        <span>{t('settings.runtimeConfig.securityHint')}</span>
                    </CardContent>
                </Card>
            )}
        </div>
    );
}

function RuntimeConfigField({
    dirty,
    entry,
    onChange,
    value,
}: {
    dirty: boolean;
    entry: RuntimeConfigEntry;
    onChange: (key: string, value: string) => void;
    value: string;
}) {
    const { t } = useLocale();
    const inputId = `runtime-config-${entry.key}`;

    return (
        <Card className={dirty ? 'border-primary/60' : undefined}>
            <CardHeader className="pb-3">
                <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                        <Label
                            className="font-mono text-sm"
                            htmlFor={inputId}
                        >
                            {entry.key}
                        </Label>
                        <CardDescription className="mt-1">{entry.description}</CardDescription>
                    </div>
                    <div className="flex shrink-0 items-center gap-1">
                        {entry.sensitive && <Badge variant="secondary">{t('settings.runtimeConfig.sensitive')}</Badge>}
                        {entry.configured && <CheckCircle2 className="size-4 text-emerald-600" />}
                    </div>
                </div>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
                {entry.sensitive ? (
                    <InputPassword
                        id={inputId}
                        onChange={(event) => onChange(entry.key, event.target.value)}
                        placeholder={
                            entry.configured
                                ? t('settings.runtimeConfig.keepSecretPlaceholder')
                                : t('settings.runtimeConfig.secretPlaceholder')
                        }
                        value={value}
                    />
                ) : entry.kind === 'list' ? (
                    <Textarea
                        id={inputId}
                        onChange={(event) => onChange(entry.key, event.target.value)}
                        placeholder={t('settings.runtimeConfig.listPlaceholder')}
                        value={value}
                    />
                ) : entry.kind === 'boolean' ? (
                    <select
                        className="border-input bg-background focus-visible:ring-ring flex h-9 w-full rounded-md border px-3 text-sm shadow-xs outline-none focus-visible:ring-1"
                        id={inputId}
                        onChange={(event) => onChange(entry.key, event.target.value)}
                        value={value}
                    >
                        <option value="true">{t('common.yes')}</option>
                        <option value="false">{t('common.no')}</option>
                    </select>
                ) : (
                    <Input
                        id={inputId}
                        inputMode={entry.kind === 'integer' ? 'numeric' : undefined}
                        onChange={(event) => onChange(entry.key, event.target.value)}
                        type={entry.kind === 'url' ? 'url' : entry.kind === 'integer' ? 'number' : 'text'}
                        value={value}
                    />
                )}
                <div className="text-muted-foreground flex flex-wrap items-center justify-between gap-2 text-xs">
                    <span>
                        {entry.defaultValue
                            ? t('settings.runtimeConfig.defaultValue', { value: entry.defaultValue })
                            : t('settings.runtimeConfig.emptyValue')}
                    </span>
                    {entry.restartRequired && <span>{t('settings.runtimeConfig.restartRequired')}</span>}
                </div>
            </CardContent>
        </Card>
    );
}
