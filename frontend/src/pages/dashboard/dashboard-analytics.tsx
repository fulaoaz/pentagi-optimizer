import { format } from 'date-fns';
import { enUS, zhCN } from 'date-fns/locale';
import { ChevronRight, Clock, Loader2, Wrench } from 'lucide-react';
import { memo, useCallback, useDeferredValue, useMemo, useRef, useState } from 'react';
import { Area, AreaChart, Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from 'recharts';

import type { FlowFragmentFragment, UsageStatsPeriod } from '@/graphql/types';
import type { Locale } from '@/lib/i18n';

import { ChartCard, ChartTooltip } from '@/components/dashboard';
import { FlowStatusBadge } from '@/components/icons/flow-status-badge';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import {
    useFlowsExecutionStatsByPeriodQuery,
    useFlowsQuery,
    useFlowsStatsByPeriodQuery,
    useToolcallsStatsByPeriodQuery,
    useUsageStatsByPeriodQuery,
} from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { cn } from '@/lib/utils';
import { formatCost, formatDuration, formatNumber, formatTokenCount } from '@/lib/utils/format';

const CHART_COLORS = {
    area1: 'var(--color-chart-1)',
    area2: 'var(--color-chart-2)',
    area3: 'var(--color-chart-3)',
    bar1: 'var(--color-chart-4)',
    bar2: 'var(--color-chart-5)',
};

const formatDateLabel = (dateString: string, locale: Locale): string => {
    try {
        return format(new Date(dateString), locale === 'zh-CN' ? 'M月d日' : 'MMM d', {
            locale: locale === 'zh-CN' ? zhCN : enUS,
        });
    } catch {
        return dateString;
    }
};

const axisTickStyle = { fill: 'var(--color-muted-foreground)', fontSize: 12 };

type FlowExecution = {
    flowId: string;
    flowTitle: string;
    tasks: Array<{
        subtasks: Array<{
            subtaskId: string;
            subtaskTitle: string;
            totalDurationSeconds: number;
            totalToolcallsCount: number;
        }>;
        taskId: string;
        taskTitle: string;
        totalDurationSeconds: number;
        totalToolcallsCount: number;
    }>;
    totalAssistantsCount: number;
    totalDurationSeconds: number;
    totalToolcallsCount: number;
};

export function DashboardAnalytics({ period }: { period: UsageStatsPeriod }) {
    const { locale, t } = useLocale();
    const { data: usageByPeriodData, loading: usageByPeriodLoading } = useUsageStatsByPeriodQuery({
        variables: { period },
    });
    const { data: toolcallsByPeriodData, loading: toolcallsByPeriodLoading } = useToolcallsStatsByPeriodQuery({
        variables: { period },
    });
    const { data: flowsByPeriodData, loading: flowsByPeriodLoading } = useFlowsStatsByPeriodQuery({
        variables: { period },
    });
    const { data: executionStatsData, loading: executionStatsLoading } = useFlowsExecutionStatsByPeriodQuery({
        variables: { period },
    });
    const { data: flowsData } = useFlowsQuery();

    const flowsTooltip = useChartTooltipAnimation();
    const toolcallsTooltip = useChartTooltipAnimation();
    const tokenUsageTooltip = useChartTooltipAnimation();
    const costTooltip = useChartTooltipAnimation();
    const formatLocalizedDateLabel = useCallback((value: string) => formatDateLabel(value, locale), [locale]);

    const flowsById = useMemo(() => {
        const map = new Map<string, FlowFragmentFragment>();
        (flowsData?.flows ?? []).forEach((flow) => {
            map.set(flow.id, flow);
        });

        return map;
    }, [flowsData?.flows]);

    const usageChartData = useMemo(
        () =>
            [...(usageByPeriodData?.usageStatsByPeriod ?? [])].reverse().map((item) => ({
                cacheIn: item.stats.totalUsageCacheIn,
                costIn: item.stats.totalUsageCostIn,
                costOut: item.stats.totalUsageCostOut,
                date: item.date,
                tokensIn: item.stats.totalUsageIn,
                tokensOut: item.stats.totalUsageOut,
                totalCost: item.stats.totalUsageCostIn + item.stats.totalUsageCostOut,
            })),
        [usageByPeriodData?.usageStatsByPeriod],
    );

    const toolcallsChartData = useMemo(
        () =>
            [...(toolcallsByPeriodData?.toolcallsStatsByPeriod ?? [])].reverse().map((item) => ({
                count: item.stats.totalCount,
                date: item.date,
                duration: item.stats.totalDurationSeconds,
            })),
        [toolcallsByPeriodData?.toolcallsStatsByPeriod],
    );

    const flowsChartData = useMemo(
        () =>
            [...(flowsByPeriodData?.flowsStatsByPeriod ?? [])].reverse().map((item) => ({
                assistants: item.stats.totalAssistantsCount,
                date: item.date,
                flows: item.stats.totalFlowsCount,
                subtasks: item.stats.totalSubtasksCount,
                tasks: item.stats.totalTasksCount,
            })),
        [flowsByPeriodData?.flowsStatsByPeriod],
    );

    const executionStats = executionStatsData?.flowsExecutionStatsByPeriod ?? [];
    // Defer the heavy 130+ row list so a period switch can repaint the charts
    // first and run the long list reconciliation as a low-priority follow-up.
    // Combined with content-visibility: auto on each row this drops the
    // pointerdown→paint INP from ~430ms into the "Good" (<200ms) zone.
    const deferredExecutionStats = useDeferredValue(executionStats);

    return (
        <div className="flex flex-col gap-6">
            <ChartCard
                description={t('dashboard.flowsChartDescription')}
                empty={!flowsByPeriodLoading && flowsChartData.length === 0}
                height={320}
                loading={flowsByPeriodLoading}
                title={t('dashboard.flowsChartTitle')}
            >
                <BarChart
                    data={flowsChartData}
                    onMouseEnter={flowsTooltip.onMouseEnter}
                    onMouseLeave={flowsTooltip.onMouseLeave}
                >
                    <CartesianGrid
                        className="stroke-border"
                        strokeDasharray="3 3"
                    />
                    <XAxis
                        dataKey="date"
                        tick={axisTickStyle}
                        tickFormatter={formatLocalizedDateLabel}
                        tickMargin={8}
                    />
                    <YAxis
                        tick={axisTickStyle}
                        tickMargin={8}
                    />
                    <Tooltip
                        content={
                            <ChartTooltip
                                labelFormatter={formatLocalizedDateLabel}
                                onFirstActive={flowsTooltip.onFirstActive}
                                sessionKey={flowsTooltip.sessionKey}
                            />
                        }
                        cursor={{ fill: 'var(--color-muted-foreground)', fillOpacity: 0.1 }}
                        isAnimationActive={flowsTooltip.isAnimationActive}
                    />
                    <Bar
                        dataKey="flows"
                        fill={CHART_COLORS.area1}
                        name={t('title.flows')}
                        radius={[4, 4, 0, 0]}
                    />
                    <Bar
                        dataKey="tasks"
                        fill={CHART_COLORS.area2}
                        name={t('flow.dashboard.tasks')}
                        radius={[4, 4, 0, 0]}
                    />
                    <Bar
                        dataKey="subtasks"
                        fill={CHART_COLORS.area3}
                        name={t('dashboard.subtasks')}
                        radius={[4, 4, 0, 0]}
                    />
                </BarChart>
            </ChartCard>

            <div className="grid gap-6 lg:grid-cols-2">
                <ChartCard
                    description={t('dashboard.toolCallsChartDescription')}
                    empty={!toolcallsByPeriodLoading && toolcallsChartData.length === 0}
                    loading={toolcallsByPeriodLoading}
                    title={t('dashboard.toolCallsChartTitle')}
                >
                    <BarChart
                        data={toolcallsChartData}
                        onMouseEnter={toolcallsTooltip.onMouseEnter}
                        onMouseLeave={toolcallsTooltip.onMouseLeave}
                    >
                        <CartesianGrid
                            className="stroke-border"
                            strokeDasharray="3 3"
                        />
                        <XAxis
                            dataKey="date"
                            tick={axisTickStyle}
                            tickFormatter={formatLocalizedDateLabel}
                            tickMargin={8}
                        />
                        <YAxis
                            tick={axisTickStyle}
                            tickMargin={8}
                        />
                        <Tooltip
                            content={
                                <ChartTooltip
                                    labelFormatter={formatLocalizedDateLabel}
                                    onFirstActive={toolcallsTooltip.onFirstActive}
                                    sessionKey={toolcallsTooltip.sessionKey}
                                />
                            }
                            cursor={{ fill: 'var(--color-muted-foreground)', fillOpacity: 0.1 }}
                            isAnimationActive={toolcallsTooltip.isAnimationActive}
                        />
                        <Bar
                            dataKey="count"
                            fill={CHART_COLORS.bar1}
                            name={t('flow.dashboard.toolCalls')}
                            radius={[4, 4, 0, 0]}
                        />
                    </BarChart>
                </ChartCard>

                <ChartCard
                    description={t('dashboard.tokensChartDescription')}
                    empty={!usageByPeriodLoading && usageChartData.length === 0}
                    loading={usageByPeriodLoading}
                    title={t('dashboard.tokensChartTitle')}
                >
                    <AreaChart
                        data={usageChartData}
                        onMouseEnter={tokenUsageTooltip.onMouseEnter}
                        onMouseLeave={tokenUsageTooltip.onMouseLeave}
                    >
                        <CartesianGrid
                            className="stroke-border"
                            strokeDasharray="3 3"
                        />
                        <XAxis
                            dataKey="date"
                            tick={axisTickStyle}
                            tickFormatter={formatLocalizedDateLabel}
                            tickMargin={8}
                        />
                        <YAxis
                            tick={axisTickStyle}
                            tickFormatter={formatTokenCount}
                            tickMargin={8}
                        />
                        <Tooltip
                            content={
                                <ChartTooltip
                                    formatter={(value) => formatTokenCount(value)}
                                    labelFormatter={formatLocalizedDateLabel}
                                    onFirstActive={tokenUsageTooltip.onFirstActive}
                                    sessionKey={tokenUsageTooltip.sessionKey}
                                />
                            }
                            isAnimationActive={tokenUsageTooltip.isAnimationActive}
                        />
                        <Area
                            dataKey="tokensIn"
                            fill={CHART_COLORS.area1}
                            fillOpacity={0.3}
                            name={t('flow.dashboard.tokensIn')}
                            stroke={CHART_COLORS.area1}
                            type="monotone"
                        />
                        <Area
                            dataKey="tokensOut"
                            fill={CHART_COLORS.area2}
                            fillOpacity={0.3}
                            name={t('flow.dashboard.tokensOut')}
                            stroke={CHART_COLORS.area2}
                            type="monotone"
                        />
                    </AreaChart>
                </ChartCard>
            </div>

            <ChartCard
                description={t('dashboard.costChartDescription')}
                empty={!usageByPeriodLoading && usageChartData.length === 0}
                height={240}
                loading={usageByPeriodLoading}
                title={t('dashboard.costChartTitle')}
            >
                <AreaChart
                    data={usageChartData}
                    onMouseEnter={costTooltip.onMouseEnter}
                    onMouseLeave={costTooltip.onMouseLeave}
                >
                    <CartesianGrid
                        className="stroke-border"
                        strokeDasharray="3 3"
                    />
                    <XAxis
                        dataKey="date"
                        tick={axisTickStyle}
                        tickFormatter={formatLocalizedDateLabel}
                        tickMargin={8}
                    />
                    <YAxis
                        tick={axisTickStyle}
                        tickFormatter={(value) => formatCost(value)}
                        tickMargin={8}
                    />
                    <Tooltip
                        content={
                            <ChartTooltip
                                formatter={(value) => formatCost(value)}
                                labelFormatter={formatLocalizedDateLabel}
                                onFirstActive={costTooltip.onFirstActive}
                                sessionKey={costTooltip.sessionKey}
                            />
                        }
                        isAnimationActive={costTooltip.isAnimationActive}
                    />
                    <Area
                        dataKey="costIn"
                        fill={CHART_COLORS.area1}
                        fillOpacity={0.3}
                        name={t('flow.dashboard.costIn')}
                        stroke={CHART_COLORS.area1}
                        type="monotone"
                    />
                    <Area
                        dataKey="costOut"
                        fill={CHART_COLORS.area3}
                        fillOpacity={0.3}
                        name={t('flow.dashboard.costOut')}
                        stroke={CHART_COLORS.area3}
                        type="monotone"
                    />
                </AreaChart>
            </ChartCard>

            <Card>
                <CardHeader>
                    <CardTitle>{t('dashboard.executionDetails')}</CardTitle>
                    <CardDescription>{t('dashboard.executionDetailsDescription')}</CardDescription>
                </CardHeader>
                <CardContent>
                    {executionStatsLoading ? (
                        <div className="flex items-center justify-center py-8">
                            <Loader2 className="text-muted-foreground size-6 animate-spin" />
                        </div>
                    ) : !deferredExecutionStats.length ? (
                        <p className="text-muted-foreground py-8 text-center text-sm">{t('dashboard.noExecutions')}</p>
                    ) : (
                        <div
                            className={cn(
                                'space-y-1 transition-opacity',
                                deferredExecutionStats !== executionStats && 'opacity-60',
                            )}
                        >
                            {deferredExecutionStats.map((flow) => (
                                <FlowExecutionItem
                                    flow={flow}
                                    flowMeta={flowsById.get(flow.flowId)}
                                    key={flow.flowId}
                                />
                            ))}
                        </div>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

const FlowExecutionItem = memo(function FlowExecutionItem({
    flow,
    flowMeta,
}: {
    flow: FlowExecution;
    flowMeta?: FlowFragmentFragment;
}) {
    const { locale, t } = useLocale();
    const [isOpen, setIsOpen] = useState(false);
    const taskCount = flow.tasks.length;
    const subtaskCount = flow.tasks.reduce((sum, task) => sum + task.subtasks.length, 0);
    const countSummary = [
        t(taskCount === 1 ? 'dashboard.taskCountOne' : 'dashboard.taskCount', { count: taskCount }),
        subtaskCount > 0
            ? t(subtaskCount === 1 ? 'dashboard.subtaskCountOne' : 'dashboard.subtaskCount', {
                  count: subtaskCount,
              })
            : null,
        flow.totalAssistantsCount > 0
            ? t(flow.totalAssistantsCount === 1 ? 'dashboard.assistantCountOne' : 'dashboard.assistantCount', {
                  count: flow.totalAssistantsCount,
              })
            : null,
    ]
        .filter(Boolean)
        .join(' · ');

    return (
        <Collapsible
            // content-visibility: auto lets the browser skip layout & paint for
            // rows outside the viewport. With ~130+ rows this is the cheapest
            // possible virtualization — no deps, no measurement, no scroll math.
            className="[contain-intrinsic-size:auto_56px] [content-visibility:auto]"
            onOpenChange={setIsOpen}
            open={isOpen}
        >
            <CollapsibleTrigger className="hover:bg-muted/50 group flex w-full items-start gap-3 rounded-lg px-3 py-2 text-left transition-colors">
                <ChevronRight className={`mt-1 size-4 shrink-0 transition-transform ${isOpen ? 'rotate-90' : ''}`} />
                <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                        <span className="truncate font-medium">
                            {flow.flowTitle || t('flow.page.untitledNumbered', { id: flow.flowId })}
                        </span>
                        {flowMeta?.status && <FlowStatusBadge status={flowMeta.status} />}
                        {flowMeta?.provider?.name && <Badge variant="secondary">{flowMeta.provider.name}</Badge>}
                    </div>
                    <div className="text-muted-foreground mt-0.5 text-xs">{countSummary}</div>
                </div>
                <div className="text-muted-foreground flex shrink-0 items-center gap-4 pt-1 text-sm">
                    <span className="flex items-center gap-1">
                        <Clock className="size-3" />
                        {formatDuration(flow.totalDurationSeconds, locale)}
                    </span>
                    <span className="flex items-center gap-1">
                        <Wrench className="size-3" />
                        {formatNumber(flow.totalToolcallsCount)}
                    </span>
                </div>
            </CollapsibleTrigger>
            <CollapsibleContent>
                <div className="ml-7 space-y-1 border-l pl-3">
                    {flow.tasks.map((task) => (
                        <TaskExecutionItem
                            key={task.taskId}
                            task={task}
                        />
                    ))}
                </div>
            </CollapsibleContent>
        </Collapsible>
    );
});

const TaskExecutionItem = memo(function TaskExecutionItem({ task }: { task: FlowExecution['tasks'][number] }) {
    const { locale, t } = useLocale();
    const [isOpen, setIsOpen] = useState(false);
    const hasSubtasks = task.subtasks.length > 0;

    return (
        <Collapsible
            onOpenChange={setIsOpen}
            open={isOpen}
        >
            <CollapsibleTrigger
                className="hover:bg-muted/50 flex w-full items-center gap-3 rounded-lg px-3 py-1.5 text-left text-sm transition-colors"
                disabled={!hasSubtasks}
            >
                {hasSubtasks ? (
                    <ChevronRight className={`size-3 shrink-0 transition-transform ${isOpen ? 'rotate-90' : ''}`} />
                ) : (
                    <span className="size-3 shrink-0" />
                )}
                <div className="text-muted-foreground flex-1 truncate">
                    {task.taskTitle || t('dashboard.taskNumbered', { id: task.taskId })}
                </div>
                <div className="text-muted-foreground flex items-center gap-4 text-xs">
                    <span className="flex items-center gap-1">
                        <Clock className="size-3" />
                        {formatDuration(task.totalDurationSeconds, locale)}
                    </span>
                    <span className="flex items-center gap-1">
                        <Wrench className="size-3" />
                        {formatNumber(task.totalToolcallsCount)}
                    </span>
                </div>
            </CollapsibleTrigger>
            {hasSubtasks && (
                <CollapsibleContent>
                    <div className="ml-6 space-y-0.5 border-l pl-3">
                        {task.subtasks.map((subtask) => (
                            <div
                                className="text-muted-foreground flex items-center gap-3 px-3 py-1 text-xs"
                                key={subtask.subtaskId}
                            >
                                <div className="flex-1 truncate">
                                    {subtask.subtaskTitle || t('dashboard.subtaskNumbered', { id: subtask.subtaskId })}
                                </div>
                                <div className="flex items-center gap-4">
                                    <span className="flex items-center gap-1">
                                        <Clock className="size-3" />
                                        {formatDuration(subtask.totalDurationSeconds, locale)}
                                    </span>
                                    <span className="flex items-center gap-1">
                                        <Wrench className="size-3" />
                                        {formatNumber(subtask.totalToolcallsCount)}
                                    </span>
                                </div>
                            </div>
                        ))}
                    </div>
                </CollapsibleContent>
            )}
        </Collapsible>
    );
});

// Disables tooltip animation on chart entry and re-enables it only after the
// tooltip has rendered at the correct position for the first time in the session.
// This prevents the tooltip from flying from (0,0) to the cursor on entry,
// while keeping smooth follow animation for all subsequent movements.
function useChartTooltipAnimation() {
    const [isAnimationActive, setIsAnimationActive] = useState(false);
    const [sessionKey, setSessionKey] = useState(0);
    const rafRef = useRef<number | undefined>(undefined);

    const onMouseEnter = () => {
        cancelAnimationFrame(rafRef.current!);
        setIsAnimationActive(false);
        setSessionKey((k) => k + 1);
    };

    const onMouseLeave = () => {
        cancelAnimationFrame(rafRef.current!);
        setIsAnimationActive(false);
    };

    // Called by ChartTooltip the first time it becomes visible in this session.
    // One rAF ensures the tooltip has painted at its initial position before
    // animation is re-enabled for subsequent cursor movements.
    const onFirstActive = () => {
        cancelAnimationFrame(rafRef.current!);
        rafRef.current = requestAnimationFrame(() => setIsAnimationActive(true));
    };

    return { isAnimationActive, onFirstActive, onMouseEnter, onMouseLeave, sessionKey };
}
