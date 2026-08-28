import { Activity, CircleDollarSign, Cpu, GitFork, Loader2 } from 'lucide-react';
import { useMemo } from 'react';

import type { UsageStatsFragmentFragment } from '@/graphql/types';

import { MetricCard } from '@/components/dashboard';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import FlowAgentIcon from '@/features/flows/agents/flow-agent-icon';
import {
    AgentType,
    useFlowStatsByFlowQuery,
    useToolcallsStatsByFlowQuery,
    useToolcallsStatsByFunctionForFlowQuery,
    useUsageStatsByAgentTypeForFlowQuery,
    useUsageStatsByFlowQuery,
    useUsageStatsByModelAgentsForFlowQuery,
} from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { translateAgentName } from '@/lib/i18n/settings-labels';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

export function FlowDashboardOverview({ flowId }: { flowId: string }) {
    const { locale, t } = useLocale();
    const { data: usageData, loading: usageLoading } = useUsageStatsByFlowQuery({
        variables: { flowId },
    });
    const { data: usageByAgentData, loading: usageByAgentLoading } = useUsageStatsByAgentTypeForFlowQuery({
        variables: { flowId },
    });
    const { data: usageByModelAgentsData, loading: usageByModelAgentsLoading } = useUsageStatsByModelAgentsForFlowQuery(
        {
            variables: { flowId },
        },
    );
    const { data: toolcallsData, loading: toolcallsLoading } = useToolcallsStatsByFlowQuery({
        variables: { flowId },
    });
    const { data: toolcallsByFunctionData, loading: toolcallsByFunctionLoading } =
        useToolcallsStatsByFunctionForFlowQuery({
            variables: { flowId },
        });
    const { data: flowStatsData, loading: flowStatsLoading } = useFlowStatsByFlowQuery({
        variables: { flowId },
    });

    const usage = usageData?.usageStatsByFlow;
    const toolcalls = toolcallsData?.toolcallsStatsByFlow;
    const flowStats = flowStatsData?.flowStatsByFlow;

    const totalCost = usage ? usage.totalUsageCostIn + usage.totalUsageCostOut : 0;
    const totalTokens = usage ? usage.totalUsageIn + usage.totalUsageOut : 0;

    const agentTypeRows = useMemo(() => {
        const seen = new Set<string>();

        return (usageByAgentData?.usageStatsByAgentTypeForFlow ?? [])
            .filter((item) => {
                if (seen.has(item.agentType)) {
                    return false;
                }

                seen.add(item.agentType);

                return true;
            })
            .map((item) => ({
                label: item.agentType,
                stats: item.stats,
            }));
    }, [usageByAgentData]);

    const modelAgentRows = useMemo(() => {
        const seen = new Set<string>();

        return (usageByModelAgentsData?.usageStatsByModelAgentsForFlow ?? []).filter((item) => {
            const key = `${item.model}|${item.provider}`;

            if (seen.has(key)) {
                return false;
            }

            seen.add(key);

            return true;
        });
    }, [usageByModelAgentsData]);

    const toolcallsByFunction = useMemo(() => {
        const seen = new Set<string>();

        return [...(toolcallsByFunctionData?.toolcallsStatsByFunctionForFlow ?? [])]
            .filter((item) => {
                if (seen.has(item.functionName)) {
                    return false;
                }

                seen.add(item.functionName);

                return true;
            })
            .sort((a, b) => b.totalCount - a.totalCount);
    }, [toolcallsByFunctionData]);

    const anyLoading = usageLoading || toolcallsLoading || flowStatsLoading;

    return (
        <div className="space-y-6">
            <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
                <MetricCard
                    description={t('flow.dashboard.taskBreakdown', {
                        assistants: flowStats?.totalAssistantsCount ?? 0,
                        subtasks: flowStats?.totalSubtasksCount ?? 0,
                    })}
                    icon={<GitFork className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title={t('flow.dashboard.tasks')}
                    value={flowStats ? formatNumber(flowStats.totalTasksCount) : '0'}
                />
                <MetricCard
                    description={t('flow.dashboard.duration', {
                        duration: toolcalls ? formatDuration(toolcalls.totalDurationSeconds, locale) : '—',
                    })}
                    icon={<Activity className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title={t('flow.dashboard.toolCalls')}
                    value={toolcalls ? formatNumber(toolcalls.totalCount) : '0'}
                />
                <MetricCard
                    description={t('flow.dashboard.tokensDescription')}
                    icon={<Cpu className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title={t('flow.dashboard.tokens')}
                    value={formatTokenCount(totalTokens)}
                />
                <MetricCard
                    description={t('flow.dashboard.costDescription')}
                    icon={<CircleDollarSign className="text-muted-foreground size-4" />}
                    loading={anyLoading}
                    title={t('flow.dashboard.cost')}
                    value={formatCost(totalCost)}
                />
            </div>

            {!!modelAgentRows.length && (
                <Card>
                    <CardHeader>
                        <CardTitle>{t('flow.dashboard.usageByModel')}</CardTitle>
                        <CardDescription>{t('flow.dashboard.usageByModelDescription')}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {usageByModelAgentsLoading ? (
                            <LoadingTable />
                        ) : (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead className="whitespace-nowrap">{t('flow.dashboard.model')}</TableHead>
                                        <TableHead className="whitespace-nowrap">
                                            {t('flow.dashboard.provider')}
                                        </TableHead>
                                        <TableHead className="whitespace-nowrap">
                                            {t('flow.dashboard.agents')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.tokensIn')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.tokensOut')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.cacheIn')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.cacheOut')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.costIn')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.costOut')}
                                        </TableHead>
                                        <TableHead className="text-right whitespace-nowrap">
                                            {t('flow.dashboard.totalCost')}
                                        </TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {modelAgentRows.map((row) => (
                                        <TableRow key={`${row.model}|${row.provider}`}>
                                            <TableCell className="font-medium">{row.model}</TableCell>
                                            <TableCell>{row.provider}</TableCell>
                                            <TableCell>
                                                <div className="flex flex-wrap gap-1">
                                                    {row.agentTypes.map((agentType) => (
                                                        <FlowAgentIcon
                                                            className="size-3.5"
                                                            key={agentType}
                                                            tooltip={translateAgentName(agentType, t)}
                                                            type={agentType as AgentType}
                                                        />
                                                    ))}
                                                </div>
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatTokenCount(row.stats.totalUsageIn)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatTokenCount(row.stats.totalUsageOut)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatTokenCount(row.stats.totalUsageCacheIn)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatTokenCount(row.stats.totalUsageCacheOut)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatCost(row.stats.totalUsageCostIn)}
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatCost(row.stats.totalUsageCostOut)}
                                            </TableCell>
                                            <TableCell className="text-right font-semibold">
                                                {formatCost(row.stats.totalUsageCostIn + row.stats.totalUsageCostOut)}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                </TableBody>
                            </Table>
                        )}
                    </CardContent>
                </Card>
            )}

            {!!agentTypeRows.length && (
                <Card>
                    <CardHeader>
                        <CardTitle>{t('flow.dashboard.usageByAgent')}</CardTitle>
                        <CardDescription>{t('flow.dashboard.usageByAgentDescription')}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {usageByAgentLoading ? (
                            <LoadingTable />
                        ) : (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>{t('flow.dashboard.agentType')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.tokensIn')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.tokensOut')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.cacheIn')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.cacheOut')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.costIn')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.costOut')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.totalCost')}</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {agentTypeRows.map((row) => (
                                        <UsageStatsRow
                                            key={row.label}
                                            label={translateAgentName(row.label, t)}
                                            stats={row.stats}
                                        />
                                    ))}
                                </TableBody>
                            </Table>
                        )}
                    </CardContent>
                </Card>
            )}

            {!!toolcallsByFunction.length && (
                <Card>
                    <CardHeader>
                        <CardTitle>{t('flow.dashboard.toolCallsByFunction')}</CardTitle>
                        <CardDescription>{t('flow.dashboard.toolCallsByFunctionDescription')}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        {toolcallsByFunctionLoading ? (
                            <LoadingTable />
                        ) : (
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>{t('flow.dashboard.function')}</TableHead>
                                        <TableHead>{t('flow.dashboard.type')}</TableHead>
                                        <TableHead className="text-right">{t('flow.dashboard.count')}</TableHead>
                                        <TableHead className="text-right">
                                            {t('flow.dashboard.totalDuration')}
                                        </TableHead>
                                        <TableHead className="text-right">
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
                                                    {item.isAgent
                                                        ? t('flow.dashboard.agent')
                                                        : t('flow.dashboard.tool')}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="text-right">
                                                {formatNumber(item.totalCount)}
                                            </TableCell>
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
            )}
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
