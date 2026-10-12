import type { ColumnDef } from '@tanstack/react-table';

import { useMutation } from '@apollo/client/react';
import { enUS, zhCN } from 'date-fns/locale';
import { Ellipsis, Eye, GitFork, Loader2, Pause, Pencil, PencilLine, Plus, Star, Trash } from 'lucide-react';
import { CheckCircle2, XCircle } from 'lucide-react';
import { useCallback, useMemo, useRef, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import {
    AppHeader,
    AppHeaderAction,
    AppHeaderActions,
    AppHeaderContent,
    AppHeaderTitle,
} from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { HeaderButton } from '@/components/shared/header-button';
import { InlineEditInput } from '@/components/shared/inline-edit';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import { ContextMenuItem, ContextMenuSeparator } from '@/components/ui/context-menu';
import { DataTable, DataTableColumnHeader } from '@/components/ui/data-table';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { StatusCard } from '@/components/ui/status-card';
import { Toggle } from '@/components/ui/toggle';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { FlowProviderLabel } from '@/features/flows/flow-provider-label';
import { RenameFlowDocument, ResultType, StatusType, type TerminalFragmentFragment } from '@/graphql/types';
import { useIsOwnFlow } from '@/hooks/use-is-own-flow';
import { useTableState } from '@/hooks/use-table-state';
import { mergeHrefWithSearchParams } from '@/lib/url-params';
import { formatTableDate } from '@/lib/utils/format';
import { useFavorites } from '@/providers/favorites-provider';
import { type Flow, useFlows } from '@/providers/flows-provider';
import { useProviders } from '@/providers/providers-provider';

const statusConfig: Record<StatusType, { variant: 'default' | 'destructive' | 'outline' | 'secondary' }> = {
    [StatusType.Created]: {
        variant: 'outline',
    },
    [StatusType.Failed]: {
        label: 'Failed',
        variant: 'outline',
    },
    [StatusType.Finished]: {
        variant: 'secondary',
    },
    [StatusType.Running]: {
        label: 'Running',
        variant: 'outline',
    },
    [StatusType.Waiting]: {
        variant: 'outline',
    },
};

function Flows() {
    const { locale, t } = useLocale();
    const dateLocale = locale === 'zh-CN' ? zhCN : enUS;
    const navigate = useNavigate();
    const location = useLocation();
    const { deleteFlow, finishFlow, flows, flowsError, isLoading, refetch } = useFlows();
    const { canToggleFavorite, isFavoriteFlow, toggleFavoriteFlow } = useFavorites();
    const { providers } = useProviders();
    const isOwnFlow = useIsOwnFlow();
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingFlow, setDeletingFlow] = useState<Flow | null>(null);
    const [finishingFlowIds, setFinishingFlowIds] = useState<Set<string>>(new Set());
    const [deletingFlowIds, setDeletingFlowIds] = useState<Set<string>>(new Set());
    const [editingFlowId, setEditingFlowId] = useState<null | string>(null);
    const editingInputRef = useRef<HTMLInputElement>(null);
    const [renameFlowMutation, { loading: isRenameLoading }] = useMutation(RenameFlowDocument);

    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

    const handleFlowOpen = useCallback(
        (flowId: string) => {
            navigate(mergeHrefWithSearchParams(`/flows/${flowId}`, new URLSearchParams(location.search)));
        },
        [navigate, location.search],
    );

    const handleFlowDeleteDialogOpen = useCallback((flow: Flow) => {
        setDeletingFlow(flow);
        setIsDeleteDialogOpen(true);
    }, []);

    const handleFlowRenameStart = useCallback((flow: Flow) => {
        setEditingFlowId(flow.id);
    }, []);

    const handleFlowDelete = async () => {
        if (!deletingFlow) {
            return;
        }

        setDeletingFlowIds((previousIds) => new Set(previousIds).add(deletingFlow.id));

        try {
            const success = await deleteFlow(deletingFlow);

            if (success) {
                setDeletingFlow(null);
            }
        } finally {
            setDeletingFlowIds((previousIds) => {
                const newIds = new Set(previousIds);
                newIds.delete(deletingFlow.id);

                return newIds;
            });
        }
    };

    const handleFlowRenameSave = useCallback(async () => {
        const newTitle = editingInputRef.current?.value.trim();

        if (!editingFlowId || !newTitle) {
            return;
        }

        try {
            const { data } = await renameFlowMutation({
                variables: {
                    flowId: editingFlowId,
                    title: newTitle,
                },
            });

            if (data?.renameFlow === ResultType.Success) {
                toast.success(t('flow.page.renamed'));
                setEditingFlowId(null);
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : t('flow.page.renameFailed');
            toast.error(errorMessage);
        }
    }, [editingFlowId, renameFlowMutation, t]);

    const handleFlowRenameCancel = useCallback(() => {
        setEditingFlowId(null);
    }, []);

    const handleFlowFinish = useCallback(
        async (flow: Flow) => {
            setFinishingFlowIds((previousIds) => new Set(previousIds).add(flow.id));

            try {
                await finishFlow(flow);
            } finally {
                setFinishingFlowIds((previousIds) => {
                    const newIds = new Set(previousIds);
                    newIds.delete(flow.id);

                    return newIds;
                });
            }
        },
        [finishFlow],
    );

    const columns: ColumnDef<Flow>[] = useMemo(
        () => [
            {
                accessorKey: 'id',
                cell: ({ row }) => <div className="font-mono text-sm whitespace-nowrap">{row.getValue('id')}</div>,
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.id')}
                    />
                ),
                maxSize: 110,
                meta: { searchable: true },
                minSize: 80,
                size: 90,
            },
            {
                accessorKey: 'title',
                cell: ({ row }) => {
                    const flow = row.original;
                    const isEditing = editingFlowId === flow.id;
                    const title = row.getValue('title') as string;

                    if (isEditing) {
                        return (
                            <div onClick={(e) => e.stopPropagation()}>
                                <InlineEditInput
                                    autoFocus
                                    busy={isRenameLoading}
                                    defaultValue={title}
                                    inputRef={editingInputRef}
                                    onCancel={handleFlowRenameCancel}
                                    onSave={handleFlowRenameSave}
                                    placeholder={t('flow.page.flowTitle')}
                                />
                            </div>
                        );
                    }

                    return <div className="truncate font-medium">{title}</div>;
                },
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.title')}
                    />
                ),
                meta: { searchable: true },
                minSize: 200,
                size: 240,
            },
            {
                accessorKey: 'status',
                cell: ({ row }) => {
                    const status = row.getValue('status') as StatusType;
                    const config = statusConfig[status];

                    return (
                        <Badge variant={config.variant}>
                            <FlowStatusIcon
                                className="size-3"
                                status={status}
                            />
                            {t(`flow.status.${status}`)}
                        </Badge>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.status')}
                    />
                ),
                maxSize: 130,
                meta: { searchable: true },
                minSize: 80,
                size: 100,
            },
            {
                // accessorFn returns the provider name as a plain string so it
                // participates in the DataTable global filter (search input).
                // The cell renderer still reads the original provider object
                // directly through `row.original`, so the icon + label stay
                // intact.
                accessorFn: (row) => row.provider?.name ?? '',
                cell: ({ row }) => (
                    <FlowProviderLabel
                        isOwnFlow={isOwnFlow(row.original)}
                        provider={row.original.provider}
                        providers={providers}
                    />
                ),
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.provider')}
                    />
                ),
                id: 'provider',
                maxSize: 240,
                meta: { searchable: true },
                minSize: 100,
                size: 160,
                sortingFn: (rowA, rowB) => {
                    const nameA = rowA.original.provider?.name || '';
                    const nameB = rowB.original.provider?.name || '';

                    return nameA.localeCompare(nameB);
                },
            },
            {
                // accessorFn joins all terminal images into one string for the
                // global search; the cell still derives its presentation from
                // the original array on `row.original`, and sortingFn keeps
                // ordering by count (more intuitive than alphabetical).
                accessorFn: (row) => (row.terminals ?? []).map((t) => t.image).join(' '),
                cell: ({ row }) => {
                    const flow = row.original;
                    const terminals = flow.terminals || [];

                    if (terminals.length === 0) {
                        return <span className="text-muted-foreground text-sm">{t('flow.list.noTerminals')}</span>;
                    }

                    const isAnyConnected = terminals.some((t: TerminalFragmentFragment) => t.connected);
                    const images = [...new Set(terminals.map((t: TerminalFragmentFragment) => t.image))];

                    return (
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <div className="flex items-center gap-2 overflow-hidden">
                                    {isAnyConnected ? (
                                        <CheckCircle2 className="size-4 shrink-0 text-green-500" />
                                    ) : (
                                        <XCircle className="text-muted-foreground size-4 shrink-0" />
                                    )}
                                    <span className="truncate text-sm">{images.join(', ')}</span>
                                </div>
                            </TooltipTrigger>
                            <TooltipContent>
                                <div className="flex flex-col gap-1">
                                    {terminals.map((terminal: TerminalFragmentFragment) => (
                                        <div
                                            className="flex items-center gap-2"
                                            key={terminal.id}
                                        >
                                            <span className="text-xs">{terminal.image}</span>
                                            <span className="text-muted-foreground text-xs">
                                                {t('flow.list.connectionStatus', {
                                                    status: terminal.connected
                                                        ? t('flow.list.connected')
                                                        : t('flow.list.disconnected'),
                                                })}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            </TooltipContent>
                        </Tooltip>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.terminals')}
                    />
                ),
                id: 'terminals',
                maxSize: 220,
                meta: { searchable: true },
                minSize: 140,
                size: 160,
                sortingFn: (rowA, rowB) => {
                    const terminalsA = rowA.original.terminals || [];
                    const terminalsB = rowB.original.terminals || [];

                    return terminalsA.length - terminalsB.length;
                },
            },
            {
                accessorKey: 'createdAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('createdAt') as string;

                    return <div className="text-sm whitespace-nowrap">{formatTableDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.created')}
                    />
                ),
                maxSize: 170,
                meta: { columnMenuLabel: 'Created' },
                minSize: 128,
                size: 136,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('createdAt') as string);
                    const dateB = new Date(rowB.getValue('createdAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                accessorKey: 'updatedAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('updatedAt') as string;

                    return <div className="text-sm whitespace-nowrap">{formatTableDate(new Date(dateString))}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('flow.list.updated')}
                    />
                ),
                maxSize: 170,
                meta: { columnMenuLabel: 'Updated' },
                minSize: 128,
                size: 136,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('updatedAt') as string);
                    const dateB = new Date(rowB.getValue('updatedAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                cell: ({ row }) => {
                    const flow = row.original;
                    const isRunning = ![StatusType.Failed, StatusType.Finished].includes(flow.status);

                    return (
                        <div className="flex items-center justify-end gap-1 opacity-0 transition-opacity group-hover:opacity-100">
                            {canToggleFavorite(flow) && (
                                <Toggle
                                    aria-label="Toggle favorite"
                                    className="border-none data-[state=on]:bg-transparent data-[state=on]:*:[svg]:fill-yellow-500 data-[state=on]:*:[svg]:stroke-yellow-500"
                                    onClick={async (event) => {
                                        event.stopPropagation();
                                        await toggleFavoriteFlow(flow.id);
                                    }}
                                    pressed={isFavoriteFlow(flow.id)}
                                    size="sm"
                                    variant="outline"
                                >
                                    <Star />
                                </Toggle>
                            )}
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        aria-label={t('common.openMenu')}
                                        className="size-8 p-0"
                                        onClick={(e) => e.stopPropagation()}
                                        variant="ghost"
                                    >
                                        <Ellipsis />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-24"
                                    onClick={(e) => e.stopPropagation()}
                                >
                                    <DropdownMenuItem onClick={() => handleFlowOpen(flow.id)}>
                                        <Eye />
                                        {t('flow.list.view')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => handleFlowRenameStart(flow)}>
                                        <PencilLine className="size-3" />
                                        {t('flow.page.rename')}
                                    </DropdownMenuItem>
                                    {isRunning && (
                                        <DropdownMenuItem
                                            disabled={finishingFlowIds.has(flow.id)}
                                            onClick={() => handleFlowFinish(flow)}
                                        >
                                            {finishingFlowIds.has(flow.id) ? (
                                                <>
                                                    <Loader2 className="animate-spin" />
                                                    {t('flow.page.finishing')}
                                                </>
                                            ) : (
                                                <>
                                                    <Pause />
                                                    {t('flow.page.finish')}
                                                </>
                                            )}
                                        </DropdownMenuItem>
                                    )}
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={deletingFlowIds.has(flow.id)}
                                        onClick={() => handleFlowDeleteDialogOpen(flow)}
                                    >
                                        {deletingFlowIds.has(flow.id) ? (
                                            <>
                                                <Loader2 className="size-4 animate-spin" />
                                                {t('common.deleting')}
                                            </>
                                        ) : (
                                            <>
                                                <Trash className="size-4" />
                                                {t('common.delete')}
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>
                    );
                },
                enableHiding: false,
                header: () => null,
                id: 'actions',
                maxSize: 100,
                meta: { preventRowClick: true },
                minSize: 90,
                size: 96,
            },
        ],
        [
            canToggleFavorite,
            deletingFlowIds,
            isOwnFlow,
            providers,
            editingFlowId,
            finishingFlowIds,
            handleFlowDeleteDialogOpen,
            handleFlowFinish,
            handleFlowOpen,
            handleFlowRenameCancel,
            handleFlowRenameSave,
            handleFlowRenameStart,
            isFavoriteFlow,
            isRenameLoading,
            t,
            toggleFavoriteFlow,
        ],
    );

    const renderRowContextMenu = useCallback(
        (flow: Flow) => {
            const isRunning = ![StatusType.Failed, StatusType.Finished].includes(flow.status);

            return (
                <>
                    {canToggleFavorite(flow) && (
                        <>
                            <ContextMenuItem onClick={async () => toggleFavoriteFlow(flow.id)}>
                                <Star />
                                {isFavoriteFlow(flow.id) ? 'Remove from favorites' : 'Add to favorites'}
                            </ContextMenuItem>
                            <ContextMenuSeparator />
                        </>
                    )}
                    <ContextMenuItem onClick={() => handleFlowOpen(flow.id)}>
                        <Eye />
                        {t('flow.list.view')}
                    </ContextMenuItem>
                    <ContextMenuItem onClick={() => handleFlowRenameStart(flow)}>
                        <Pencil />
                        {t('flow.page.rename')}
                    </ContextMenuItem>

                    {isRunning && (
                        <ContextMenuItem
                            disabled={finishingFlowIds.has(flow.id)}
                            onClick={() => handleFlowFinish(flow)}
                        >
                            <Pause />
                            {finishingFlowIds.has(flow.id) ? t('flow.page.finishing') : t('flow.page.finish')}
                        </ContextMenuItem>
                    )}
                    <ContextMenuSeparator />
                    <ContextMenuItem
                        disabled={deletingFlowIds.has(flow.id)}
                        onClick={() => handleFlowDeleteDialogOpen(flow)}
                    >
                        <Trash />
                        {deletingFlowIds.has(flow.id) ? t('common.deleting') : t('common.delete')}
                    </ContextMenuItem>
                </>
            );
        },
        [
            canToggleFavorite,
            deletingFlowIds,
            finishingFlowIds,
            handleFlowDeleteDialogOpen,
            handleFlowFinish,
            handleFlowOpen,
            handleFlowRenameStart,
            isFavoriteFlow,
            t,
            toggleFavoriteFlow,
        ],
    );

    const handleRowClick = useCallback(
        (flow: Flow) => {
            if (editingFlowId !== flow.id) {
                handleFlowOpen(flow.id);
            }
        },
        [editingFlowId, handleFlowOpen],
    );

    const pageHeader = (
        <header className="bg-background sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
            <div className="flex min-w-0 flex-1 items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1 shrink-0" />
                <Separator
                    className="h-4 shrink-0"
                    orientation="vertical"
                />
                <Breadcrumb className="min-w-0 flex-1">
                    <BreadcrumbList className="min-w-0 flex-nowrap">
                        <BreadcrumbItem className="min-w-0">
                            <GitFork className="size-4 shrink-0" />
                            <BreadcrumbPage className="min-w-0 truncate">{t('flow.page.flows')}</BreadcrumbPage>
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
            </div>
            <div className="flex shrink-0 items-center gap-2 px-4">
                <HeaderButton
                    icon={<Plus />}
                    label={t('flow.list.newFlow')}
                    onClick={() => navigate('/flows/new')}
                    variant="secondary"
                />
            </div>
        </header>
    );

    if (isLoading) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        description={t('flow.list.loadingDescription')}
                        icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                        title={t('flow.list.loadingTitle')}
                    />
                </div>
            </>
        );
    }

    if (flows.length === 0) {
        return (
            <>
                {pageHeader}
                <div className="flex flex-col gap-4 p-4">
                    <StatusCard
                        action={
                            <Button
                                onClick={() => navigate('/flows/new')}
                                variant="secondary"
                            >
                                <Plus />
                                {t('flow.list.newFlow')}
                            </Button>
                        }
                        description={t('flow.list.emptyDescription')}
                        icon={<GitFork className="text-muted-foreground size-8" />}
                        title={t('flow.list.emptyTitle')}
                    />
                </div>
            </>
        );
    }

    return (
        <>
            {pageHeader}
            <div className="flex flex-col gap-4 p-4 pt-0">
                <DataTable<Flow>
                    columns={columns}
                    data={flows}
                    empty={{ entityName: t('flow.list.entityName') }}
                    filterPlaceholder={t('flow.list.filterPlaceholder')}
                    filterValue={filter}
                    isVirtualized
                    label="Flows"
                    onFilterChange={setFilter}
                    onPageChange={handlePageChange}
                    onRowClick={handleRowClick}
                    pageIndex={currentPage}
                    renderRowContextMenu={renderRowContextMenu}
                />

                <ConfirmationDialog
                    cancelText={t('common.cancel')}
                    confirmText={t('common.delete')}
                    handleConfirm={handleFlowDelete}
                    handleOpenChange={setIsDeleteDialogOpen}
                    isOpen={isDeleteDialogOpen}
                    itemName={deletingFlow?.title}
                    itemType={t('title.flow')}
                    title={t('flow.page.deleteTitle')}
                />
            </div>
        </>
    );
}

export default Flows;
