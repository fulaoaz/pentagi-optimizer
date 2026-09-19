import { useQuery } from '@apollo/client/react';
import { AlertCircle, Globe } from 'lucide-react';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { SearchEnginesStatusDocument } from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';

export function SettingsSearchEngines() {
    const { t } = useLocale();
    const { data, error, loading } = useQuery(SearchEnginesStatusDocument);

    const engines = data?.searchEnginesStatus ?? [];

    if (error) {
        return (
            <Alert variant="destructive">
                <AlertCircle className="size-4" />
                <AlertTitle>{t('settings.searchEngines.loadError')}</AlertTitle>
                <AlertDescription>{error.message}</AlertDescription>
            </Alert>
        );
    }

    return (
        <div className="mx-auto flex w-full max-w-4xl flex-col gap-4">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Globe className="size-5" />
                        {t('settings.searchEngines.title')}
                    </CardTitle>
                    <CardDescription>{t('settings.searchEngines.description')}</CardDescription>
                </CardHeader>
                <CardContent className="flex flex-col gap-3">
                    {loading &&
                        Array.from({ length: 5 }).map((_, index) => (
                            <Skeleton
                                className="h-16 w-full"
                                key={index}
                            />
                        ))}
                    {!loading &&
                        engines.map((engine) => (
                            <div
                                className="flex items-start justify-between gap-4 rounded-lg border p-3"
                                key={engine.engineType}
                            >
                                <div className="flex min-w-0 flex-col gap-1">
                                    <span className="font-medium">{engine.name}</span>
                                    <span className="text-muted-foreground text-sm">{engine.description}</span>
                                    {!engine.available && engine.missing && (
                                        <code className="bg-muted rounded px-2 py-0.5 text-xs">{engine.missing}</code>
                                    )}
                                </div>
                                <Badge variant={engine.available ? 'default' : 'secondary'}>
                                    {engine.available
                                        ? t('settings.searchEngines.enabled')
                                        : t('settings.searchEngines.disabled')}
                                </Badge>
                            </div>
                        ))}
                    {!loading && !error && engines.length === 0 && (
                        <p className="text-muted-foreground text-sm">{t('settings.searchEngines.empty')}</p>
                    )}
                </CardContent>
            </Card>
            <p className="text-muted-foreground text-xs">{t('settings.searchEngines.restartHint')}</p>
        </div>
    );
}
