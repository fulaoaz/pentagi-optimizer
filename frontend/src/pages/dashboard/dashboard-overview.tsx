import { useQuery } from '@apollo/client/react';
import { Activity, CircleDollarSign, Cpu, GitFork, Loader2 } from 'lucide-react';
import { useMemo } from 'react';

import type { UsageStatsFragmentFragment } from '@/graphql/types';

import { DashboardError, LoadingTable, MetricCard, UsageStatsRow } from '@/components/dashboard';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import {
    FlowsStatsTotalDocument,
    ToolcallsStatsByFunctionDocument,
    ToolcallsStatsTotalDocument,
    UsageStatsByAgentTypeDocument,
    UsageStatsByModelDocument,
    UsageStatsByProviderDocument,
    UsageStatsTotalDocument,
} from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { translateAgentName } from '@/lib/i18n/settings-labels';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

export function DashboardOverview() {
    const { locale, t } = useLocale();
    const {
        data: usageTotalData,
        error: usageTotalDataError,
        loading: usageTotalLoading,
    } = useQuery(UsageStatsTotalDocument);
    const { data: usageByProviderData, loading: usageByProviderLoading } = useQuery(UsageStatsByProviderDocument);
    const { data: usageByModelData, loading: usageByModelLoading } = useQuery(UsageStatsByModelDocument);
    const { data: usageByAgentTypeData, loading: usageByAgentTypeLoading } = useQuery(UsageStatsByAgentTypeDocument);
    const {
        data: toolcallsTotalData,
        error: toolcallsTotalDataError,
        loading: toolcallsTotalLoading,
    } = useQuery(ToolcallsStatsTotalDocument);
    const { data: toolcallsByFunctionData, loading: toolcallsByFunctionLoading } = useQuery(
        ToolcallsStatsByFunctionDocument,
    );
    const {
        data: flowsTotalData,
        error: flowsTotalDataError,
        loading: flowsTotalLoading,
    } = useQuery(FlowsStatsTotalDocument);

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
                    description={`Tasks: ${flowsTotal?.totalTasksCount ?? 0} · Subtasks: ${flowsTotal?.totalSubtasksCount ?? 0} · Assistants: ${flowsTotal?.totalAssistantsCount ?? 0}`}
                    error={!!flowsTotalError && !flowsTotalData}
                    icon={<GitFork className="text-muted-foreground size-4" />}
                    loading={flowsTotalLoading && !flowsTotalData}
                    title="Total Flows"
                    value={flowsTotal ? formatNumber(flowsTotal.totalFlowsCount) : '0'}
                />
                <MetricCard
                    description={`Total duration: ${toolcallsTotal ? formatDuration(toolcallsTotal.totalDurationSeconds) : '—'}`}
                    error={!!toolcallsTotalError && !toolcallsTotalData}
                    icon={<Activity className="text-muted-foreground size-4" />}
                    loading={toolcallsTotalLoading && !toolcallsTotalData}
                    title="Tool Calls"
                    value={toolcallsTotal ? formatNumber(toolcallsTotal.totalCount) : '0'}
                />
                <MetricCard
                    description="Input + Output tokens processed"
                    error={!!usageTotalError && !usageTotalData}
                    icon={<Cpu className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading && !usageTotalData}
                    title="Total Tokens"
                    value={formatTokenCount(totalTokens)}
                />
                <MetricCard
                    description="Total LLM spending across all providers"
                    error={!!usageTotalError && !usageTotalData}
                    icon={<CircleDollarSign className="text-muted-foreground size-4" />}
                    loading={usageTotalLoading && !usageTotalData}
                    title="Total Cost"
                    value={formatCost(totalCost)}
                />
            </div>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByProvider')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByProviderDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByProviderLoading && !usageByProviderData ? (
                        <LoadingTable />
                    ) : usageByProviderError && !usageByProviderData ? (
                        <ErrorTable />
                    ) : (
                        <UsageStatsTable
                            label="Usage by Provider"
                            rows={providerRows}
                        />
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.usageByModel')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByModelDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByModelLoading && !usageByModelData ? (
                        <LoadingTable />
                    ) : usageByModelError && !usageByModelData ? (
                        <ErrorTable />
                    ) : (
                        <UsageStatsTable
                            label="Usage by Model"
                            rows={modelRows}
                        />
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('flow.dashboard.usageByAgent')}</CardTitle>
                    <CardDescription>{t('dashboard.usageByAgentDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {usageByAgentTypeLoading && !usageByAgentTypeData ? (
                        <LoadingTable />
                    ) : usageByAgentTypeError && !usageByAgentTypeData ? (
                        <ErrorTable />
                    ) : (
                        <UsageStatsTable
                            label="Usage by Agent Type"
                            rows={agentTypeRows}
                        />
                    )}
                </CardContent>
            </Card>

            <Card>
                <CardHeader>
                    <CardTitle>{t('flow.dashboard.toolCallsByFunction')}</CardTitle>
                    <CardDescription>{t('dashboard.toolCallsByFunctionDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {toolcallsByFunctionLoading && !toolcallsByFunctionData ? (
                        <LoadingTable />
                    ) : toolcallsByFunctionError && !toolcallsByFunctionData ? (
                        <ErrorTable />
                    ) : (
                        <Table aria-label="Tool Calls by Function">
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

function ErrorTable() {
    return <DashboardError className="py-8" />;
}

function UsageStatsTable({
    label,
    rows,
}: {
    label: string;
    rows: Array<{ label: string; stats: UsageStatsFragmentFragment }>;
}) {
    return (
        <Table aria-label={label}>
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
