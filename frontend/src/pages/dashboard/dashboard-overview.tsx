import { Activity, CircleDollarSign, Cpu, GitFork, Loader2 } from 'lucide-react';
import { useMemo } from 'react';

import type { UsageStatsFragmentFragment } from '@/graphql/types';

import { MetricCard } from '@/components/dashboard';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
    useFlowsStatsTotalQuery,
    useToolcallsStatsByFunctionQuery,
    useToolcallsStatsTotalQuery,
    useUsageStatsByAgentTypeQuery,
    useUsageStatsByModelQuery,
    useUsageStatsByProviderQuery,
    useUsageStatsTotalQuery,
} from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { translateAgentName } from '@/lib/i18n/settings-labels';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

export function DashboardOverview() {
    const { locale, t } = useLocale();
    const { data: usageTotalData, loading: usageTotalLoading } = useUsageStatsTotalQuery();
    const { data: usageByProviderData, loading: usageByProviderLoading } = useUsageStatsByProviderQuery();
    const { data: usageByModelData, loading: usageByModelLoading } = useUsageStatsByModelQuery();
    const { data: usageByAgentTypeData, loading: usageByAgentTypeLoading } = useUsageStatsByAgentTypeQuery();
    const { data: toolcallsTotalData, loading: toolcallsTotalLoading } = useToolcallsStatsTotalQuery();
    const { data: toolcallsByFunctionData, loading: toolcallsByFunctionLoading } = useToolcallsStatsByFunctionQuery();
    const { data: flowsTotalData, loading: flowsTotalLoading } = useFlowsStatsTotalQuery();

    const usageTotal = usageTotalData?.usageStatsTotal;
    const toolcallsTotal = toolcallsTotalData?.toolcallsStatsTotal;
    const flowsTotal = flowsTotalData?.flowsStatsTotal;

    const totalCost = usageTotal ? usageTotal.totalUsageCostIn + usageTotal.totalUsageCostOut : 0;
    const totalTokens = usageTotal ? usageTotal.totalUsageIn + usageTotal.totalUsageOut : 0;

    const providerRows = (usageByProviderData?.usageStatsByProvider ?? []).map((item) => ({
        label: item.provider,
        stats: item.stats,
    }));
    const modelRows = (usageByModelData?.usageStatsByModel ?? []).map((item) => ({
        label: t('dashboard.modelProviderLabel', { model: item.model, provider: item.provider }),
        stats: item.stats,
    }));
    const agentTypeRows = useMemo(
        () =>
            (usageByAgentTypeData?.usageStatsByAgentType ?? []).map((item) => ({
                label: translateAgentName(item.agentType, t),
                stats: item.stats,
            })),
        [t, usageByAgentTypeData?.usageStatsByAgentType],
    );

    const toolcallsByFunction = [...(toolcallsByFunctionData?.toolcallsStatsByFunction ?? [])].sort(
        (a, b) => b.totalCount - a.totalCount,
    );

    return (
        <div className="flex flex-col gap-6">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <MetricCard
                    description={t('dashboard.flowBreakdown', {
                        assistants: flowsTotal?.totalAssistantsCount ?? 0,
                        subtasks: flowsTotal?.totalSubtasksCount ?? 0,
                        tasks: flowsTotal?.totalTasksCount ?? 0,
                    })}
                    icon={<GitFork className="text-muted-foreground size-4" />}
                    loading={flowsTotalLoading}
                    title={t('dashboard.totalFlows')}
                    value={flowsTotal ? formatNumber(flowsTotal.totalFlowsCount) : '0'}
                />
                <MetricCard
                    description={t('dashboard.totalDurationValue', {
                        duration: toolcallsTotal ? formatDuration(toolcallsTotal.totalDurationSeconds, locale) : '—',
                    })}
                    icon={<Activity className="text-muted-foreground size-4" />}
                    loading={toolcallsTotalLoading}
                    title={t('flow.dashboard.toolCalls')}
                    value={toolcallsTotal ? formatNumber(toolcallsTotal.totalCount) : '0'}
                />
                <MetricCard
                    description={t('dashboard.totalTokensDescription')}
                    icon={<Cpu className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading}
                    title={t('dashboard.totalTokens')}
                    value={formatTokenCount(totalTokens)}
                />
                <MetricCard
                    description={t('dashboard.totalCostDescription')}
                    icon={<CircleDollarSign className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading}
                    title={t('flow.dashboard.totalCost')}
                    value={formatCost(totalCost)}
                />
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByProvider')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByProviderDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByProviderLoading ? <LoadingTable /> : <UsageStatsTable rows={providerRows} />}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByModel')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByModelDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByModelLoading ? <LoadingTable /> : <UsageStatsTable rows={modelRows} />}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('flow.dashboard.usageByAgent')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByAgentDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByAgentTypeLoading ? <LoadingTable /> : <UsageStatsTable rows={agentTypeRows} />}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('flow.dashboard.toolCallsByFunction')}</CardTitle>
                    <CardDescription>{t('dashboard.toolCallsByFunctionDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {toolcallsByFunctionLoading ? (
                        <LoadingTable />
                    ) : (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead className="whitespace-nowrap">{t('flow.dashboard.function')}</TableHead>
                                    <TableHead className="whitespace-nowrap">{t('flow.dashboard.type')}</TableHead>
                                    <TableHead className="text-right whitespace-nowrap">
                                        {t('flow.dashboard.count')}
                                    </TableHead>
                                    <TableHead className="text-right whitespace-nowrap">
                                        {t('flow.dashboard.totalDuration')}
                                    </TableHead>
                                    <TableHead className="text-right whitespace-nowrap">
                                        {t('flow.dashboard.averageDuration')}
                                    </TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {toolcallsByFunction.map((item) => (
                                    <TableRow key={item.functionName}>
                                        <TableCell className="font-medium">{item.functionName}</TableCell>
                                        <TableCell>
                                            <Badge variant={item.isAgent ? 'secondary' : 'outline'}>
                                                {item.isAgent ? t('flow.dashboard.agent') : t('flow.dashboard.tool')}
                                            </Badge>
                                        </TableCell>
                                        <TableCell className="text-right">{formatNumber(item.totalCount)}</TableCell>
                                        <TableCell className="text-right">
                                            {formatDuration(item.totalDurationSeconds, locale)}
                                        </TableCell>
                                        <TableCell className="text-right">
                                            {formatDuration(item.avgDurationSeconds, locale)}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

function LoadingTable() {
    return (
        <div className="flex items-center justify-center py-8">
            <Loader2 className="text-muted-foreground size-6 animate-spin" />
        </div>
    );
}

function UsageStatsRow({ label, stats }: { label: string; stats: UsageStatsFragmentFragment }) {
    return (
        <TableRow>
            <TableCell className="font-medium">{label}</TableCell>
            <TableCell className="text-right">{formatTokenCount(stats.totalUsageIn)}</TableCell>
            <TableCell className="text-right">{formatTokenCount(stats.totalUsageOut)}</TableCell>
            <TableCell className="text-right">{formatTokenCount(stats.totalUsageCacheIn)}</TableCell>
            <TableCell className="text-right">{formatTokenCount(stats.totalUsageCacheOut)}</TableCell>
            <TableCell className="text-right">{formatCost(stats.totalUsageCostIn)}</TableCell>
            <TableCell className="text-right">{formatCost(stats.totalUsageCostOut)}</TableCell>
            <TableCell className="text-right font-semibold">
                {formatCost(stats.totalUsageCostIn + stats.totalUsageCostOut)}
            </TableCell>
        </TableRow>
    );
}

function UsageStatsTable({ rows }: { rows: Array<{ label: string; stats: UsageStatsFragmentFragment }> }) {
    const { t } = useLocale();

    return (
        <Table>
            <TableHeader>
                <TableRow>
                    <TableHead className="whitespace-nowrap">{t('dashboard.name')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.tokensIn')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.tokensOut')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.cacheIn')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.cacheOut')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.costIn')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.costOut')}</TableHead>
                    <TableHead className="text-right whitespace-nowrap">{t('flow.dashboard.totalCost')}</TableHead>
                </TableRow>
            </TableHeader>
            <TableBody>
                {rows.map((row) => (
                    <UsageStatsRow
                        key={row.label}
                        label={row.label}
                        stats={row.stats}
                    />
                ))}
            </TableBody>
        </Table>
    );
}
