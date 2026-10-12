import { useQuery } from '@apollo/client/react';
import { Check, Clipboard, ExternalLink, Network, ShieldCheck, Terminal } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { RuntimeConfigDocument } from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';

function EndpointCard({ endpoint, label }: { endpoint: string; label: string }) {
    return (
        <div className="rounded-lg border p-3">
            <div className="text-muted-foreground text-xs">{label}</div>
            <code className="mt-1 block break-all text-xs">{endpoint}</code>
        </div>
    );
}

function McpConnectors() {
    const { t } = useLocale();
    const { data } = useQuery(RuntimeConfigDocument);
    const [copied, setCopied] = useState(false);

    const config = useMemo(() => {
        const entries = data?.runtimeConfig ?? [];
        const enabled = entries.find((entry) => entry.key === 'MCP_ENABLED');
        const apiKey = entries.find((entry) => entry.key === 'MCP_API_KEY');

        return {
            apiKeyConfigured: apiKey?.configured ?? false,
            enabled: enabled ? enabled.value === 'true' : null,
        };
    }, [data?.runtimeConfig]);

    const baseUrl = `${window.location.origin}/api/v1`;
    const connectorConfig = JSON.stringify(
        {
            mcpServers: {
                pentagi: {
                    url: `${baseUrl}/mcp`,
                },
            },
        },
        null,
        2,
    );

    const handleCopy = async () => {
        await navigator.clipboard.writeText(connectorConfig);
        setCopied(true);
        window.setTimeout(() => setCopied(false), 1800);
    };

    return (
        <div className="flex flex-col gap-6 p-4">
            <header className="flex flex-wrap items-start justify-between gap-4">
                <div>
                    <h1 className="flex items-center gap-2 text-2xl font-semibold">
                        <Network className="size-6 text-primary" />
                        {t('integrations.mcp.title')}
                    </h1>
                    <p className="text-muted-foreground mt-2 max-w-3xl">{t('integrations.mcp.description')}</p>
                </div>
                <Button
                    asChild
                    variant="outline"
                >
                    <Link to="/settings/runtime-config/mcp">{t('integrations.mcp.openSettings')}</Link>
                </Button>
            </header>

            <div className="grid gap-4 lg:grid-cols-[1.35fr_1fr]">
                <Card>
                    <CardHeader>
                        <div className="flex items-start justify-between gap-4">
                            <div>
                                <CardTitle className="flex items-center gap-2">
                                    <Network className="size-5" />
                                    {t('integrations.mcp.bridgeName')}
                                </CardTitle>
                                <CardDescription>{t('integrations.mcp.bridgeDescription')}</CardDescription>
                            </div>
                            <Badge variant={config.enabled === false ? 'secondary' : 'default'}>
                                {config.enabled === false
                                    ? t('integrations.mcp.disabled')
                                    : config.enabled === null
                                      ? t('integrations.mcp.statusUnknown')
                                      : t('integrations.mcp.enabled')}
                            </Badge>
                        </div>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-5">
                        <div className="grid gap-3 sm:grid-cols-2">
                            <EndpointCard
                                endpoint={`${baseUrl}/mcp`}
                                label={t('integrations.mcp.sseEndpoint')}
                            />
                            <EndpointCard
                                endpoint={`${baseUrl}/mcp/message`}
                                label={t('integrations.mcp.messageEndpoint')}
                            />
                        </div>
                        <div className="bg-muted rounded-lg p-4">
                            <div className="mb-3 flex items-center justify-between gap-3">
                                <div className="flex items-center gap-2 text-sm font-medium">
                                    <Terminal className="size-4" />
                                    {t('integrations.mcp.configSnippet')}
                                </div>
                                <Button
                                    aria-label={t('integrations.mcp.copyConfig')}
                                    onClick={handleCopy}
                                    size="sm"
                                    variant="outline"
                                >
                                    {copied ? <Check className="size-4" /> : <Clipboard className="size-4" />}
                                    {copied ? t('integrations.mcp.copied') : t('common.copy')}
                                </Button>
                            </div>
                            <pre className="overflow-x-auto text-xs"><code>{connectorConfig}</code></pre>
                        </div>
                    </CardContent>
                </Card>

                <div className="flex flex-col gap-4">
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2 text-base">
                                <ShieldCheck className="size-5 text-emerald-600" />
                                {t('integrations.mcp.securityTitle')}
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="text-muted-foreground flex flex-col gap-3 text-sm">
                            <div className="flex items-center justify-between gap-3">
                                <span>{t('integrations.mcp.apiKeyStatus')}</span>
                                <Badge variant={config.apiKeyConfigured ? 'default' : 'secondary'}>
                                    {config.apiKeyConfigured
                                        ? t('integrations.mcp.configured')
                                        : t('integrations.mcp.notConfigured')}
                                </Badge>
                            </div>
                            <p>{t('integrations.mcp.securityDescription')}</p>
                        </CardContent>
                    </Card>
                    <Card>
                        <CardHeader>
                            <CardTitle className="text-base">{t('integrations.mcp.developerTools')}</CardTitle>
                        </CardHeader>
                        <CardContent className="flex flex-wrap gap-2">
                            <Button
                                asChild
                                size="sm"
                                variant="outline"
                            >
                                <a href="/api/v1/graphql/playground" target="_blank">
                                    {t('integrations.mcp.graphql')}
                                    <ExternalLink className="size-3" />
                                </a>
                            </Button>
                            <Button
                                asChild
                                size="sm"
                                variant="outline"
                            >
                                <a href="/api/v1/swagger/index.html" target="_blank">
                                    {t('integrations.mcp.swagger')}
                                    <ExternalLink className="size-3" />
                                </a>
                            </Button>
                        </CardContent>
                    </Card>
                </div>
            </div>

            <Alert>
                <Network className="size-4" />
                <AlertTitle>{t('integrations.mcp.howToUseTitle')}</AlertTitle>
                <AlertDescription>{t('integrations.mcp.howToUseDescription')}</AlertDescription>
            </Alert>
        </div>
    );
}

export default McpConnectors;
