import { useQuery } from '@apollo/client/react';
import { AlertCircle, Info, Search, Settings2 } from 'lucide-react';
import { useMemo, useState } from 'react';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { RuntimeConfigDocument } from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';

export default function SettingsRuntimeConfig() {
    const { t } = useLocale();
    const [query, setQuery] = useState('');
    const { data, error, loading } = useQuery(RuntimeConfigDocument);
    const entries = data?.runtimeConfig;
    const normalizedQuery = query.trim().toLowerCase();
    const groups = useMemo(() => {
        const result = new Map<string, typeof entries>();

        for (const entry of entries ?? []) {
            if (
                normalizedQuery &&
                ![entry.key, entry.category, entry.description, entry.value]
                    .join(' ')
                    .toLowerCase()
                    .includes(normalizedQuery)
            ) {
                continue;
            }

            const group = result.get(entry.category) ?? [];
            group.push(entry);
            result.set(entry.category, group);
        }

        return [...result.entries()];
    }, [entries, normalizedQuery]);
    const configuredCount = (entries ?? []).filter((entry) => entry.configured).length;

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>{t('settings.runtimeConfig.loadError')}</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    return (
        <div className="mx-auto flex w-full max-w-6xl flex-col gap-4">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Settings2 className="size-5" />
                        {t('settings.runtimeConfig.title')}
                    </CardTitle>
                    <CardDescription>{t('settings.runtimeConfig.description')}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="text-muted-foreground flex items-center gap-2 text-sm">
                        <Info className="size-4" />
                        {t('settings.runtimeConfig.summary', {
                            configured: configuredCount,
                            total: entries?.length ?? 0,
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

            {loading && <Skeleton className="h-64 w-full" />}

            {!loading && groups.length === 0 && (
                <Card>
                    <CardContent className="text-muted-foreground py-10 text-center text-sm">
                        {t('settings.runtimeConfig.empty')}
                    </CardContent>
                </Card>
            )}

            {!loading &&
                groups.map(([category, categoryEntries]) => (
                    <Card key={category}>
                        <CardHeader className="pb-3">
                            <CardTitle className="text-base">{category}</CardTitle>
                        </CardHeader>
                        <CardContent className="grid gap-3 md:grid-cols-2">
                            {categoryEntries.map((entry) => (
                                <div
                                    className="rounded-lg border p-3"
                                    key={entry.key}
                                >
                                    <div className="flex items-start justify-between gap-3">
                                        <div className="min-w-0">
                                            <code className="text-sm font-medium">{entry.key}</code>
                                            <p className="text-muted-foreground mt-1 text-xs">{entry.description}</p>
                                        </div>
                                        <Badge variant={entry.configured ? 'default' : 'secondary'}>
                                            {entry.configured
                                                ? t('settings.runtimeConfig.configured')
                                                : t('settings.runtimeConfig.notConfigured')}
                                        </Badge>
                                    </div>
                                    <div className="bg-muted mt-3 rounded px-2 py-1.5 font-mono text-xs break-all">
                                        {entry.value || t('settings.runtimeConfig.emptyValue')}
                                    </div>
                                    {entry.defaultValue && (
                                        <p className="text-muted-foreground mt-2 text-xs">
                                            {t('settings.runtimeConfig.defaultValue', { value: entry.defaultValue })}
                                        </p>
                                    )}
                                    {entry.restartRequired && (
                                        <p className="text-muted-foreground mt-1 text-xs">
                                            {t('settings.runtimeConfig.restartRequired')}
                                        </p>
                                    )}
                                </div>
                            ))}
                        </CardContent>
                    </Card>
                ))}
        </div>
    );
}
