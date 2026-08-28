import type { ColumnDef, Row } from '@tanstack/react-table';

import { enUS, zhCN } from 'date-fns/locale';
import { AlertCircle, ChevronDown, Copy, Ellipsis, Loader2, Pencil, Plus, Settings, Trash } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import type { ProviderConfigFragmentFragment } from '@/graphql/types';

import Anthropic from '@/components/icons/anthropic';
import Bedrock from '@/components/icons/bedrock';
import Custom from '@/components/icons/custom';
import DeepSeek from '@/components/icons/deepseek';
import Gemini from '@/components/icons/gemini';
import GLM from '@/components/icons/glm';
import Kimi from '@/components/icons/kimi';
import Ollama from '@/components/icons/ollama';
import OpenAi from '@/components/icons/open-ai';
import Qwen from '@/components/icons/qwen';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
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
import { StatusCard } from '@/components/ui/status-card';
import { ProviderType, useDeleteProviderMutation, useSettingsProvidersQuery } from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { useTableState } from '@/hooks/use-table-state';
import { translateAgentName, translateProviderFieldName } from '@/lib/i18n/settings-labels';
import { formatDate } from '@/lib/utils/format';
type Provider = ProviderConfigFragmentFragment;

const providerIcons: Record<ProviderType, React.ComponentType<React.SVGProps<SVGSVGElement>>> = {
    [ProviderType.Anthropic]: Anthropic,
    [ProviderType.Bedrock]: Bedrock,
    [ProviderType.Custom]: Custom,
    [ProviderType.Deepseek]: DeepSeek,
    [ProviderType.Gemini]: Gemini,
    [ProviderType.Glm]: GLM,
    [ProviderType.Kimi]: Kimi,
    [ProviderType.Ollama]: Ollama,
    [ProviderType.Openai]: OpenAi,
    [ProviderType.Qwen]: Qwen,
};

const providerTypes = [
    { label: 'Anthropic', type: ProviderType.Anthropic },
    { label: 'Bedrock', type: ProviderType.Bedrock },
    { label: 'Custom', type: ProviderType.Custom },
    { label: 'DeepSeek', type: ProviderType.Deepseek },
    { label: 'Gemini', type: ProviderType.Gemini },
    { label: 'GLM', type: ProviderType.Glm },
    { label: 'Kimi', type: ProviderType.Kimi },
    { label: 'Ollama', type: ProviderType.Ollama },
    { label: 'OpenAI', type: ProviderType.Openai },
    { label: 'Qwen', type: ProviderType.Qwen },
];

function SettingsProviders() {
    const { locale, t } = useLocale();
    const { data, error, loading: isLoading } = useSettingsProvidersQuery();
    const [deleteProvider, { error: deleteError, loading: isDeleteLoading }] = useDeleteProviderMutation();
    const [deleteErrorMessage, setDeleteErrorMessage] = useState<null | string>(null);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [deletingProvider, setDeletingProvider] = useState<null | Provider>(null);
    const navigate = useNavigate();
    const dateLocale = locale === 'zh-CN' ? zhCN : enUS;

    const { filter, pageIndex: currentPage, setFilter, setPage: handlePageChange } = useTableState();

    const handleProviderDelete = useCallback(
        async (providerId: string | undefined) => {
            if (!providerId) {
                return;
            }

            try {
                setDeleteErrorMessage(null);

                await deleteProvider({
                    refetchQueries: ['settingsProviders'],
                    variables: { providerId: providerId.toString() },
                });

                setDeletingProvider(null);
                setDeleteErrorMessage(null);
            } catch (error) {
                setDeleteErrorMessage(error instanceof Error ? error.message : t('settings.providers.deleteError'));
            }
        },
        [deleteProvider, t],
    );

    const handleProviderEdit = useCallback(
        (providerId: string) => {
            navigate(`/settings/providers/${providerId}`);
        },
        [navigate],
    );

    const handleProviderClone = useCallback(
        (providerId: string) => {
            navigate(`/settings/providers/new?id=${providerId}`);
        },
        [navigate],
    );

    const handleProviderDeleteDialogOpen = useCallback((provider: Provider) => {
        setDeletingProvider(provider);
        setIsDeleteDialogOpen(true);
    }, []);

    const columns: ColumnDef<Provider>[] = useMemo(
        () => [
            {
                accessorKey: 'name',
                cell: ({ row }) => <div className="truncate font-medium">{row.getValue('name')}</div>,
                enableHiding: false,
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('settings.providers.name')}
                    />
                ),
                // Name flexes to fill remaining width — fixed `size` would push
                // the Type column off-screen on narrow viewports (e.g. 375px).
                meta: { searchable: true },
            },
            {
                accessorKey: 'type',
                cell: ({ row }) => {
                    const providerType = row.getValue('type') as ProviderType;
                    const Icon = providerIcons[providerType];
                    const label =
                        providerType === ProviderType.Custom
                            ? t('settings.providers.custom')
                            : providerTypes.find((p) => p.type === providerType)?.label || providerType;

                    return (
                        <Badge
                            className="max-w-full whitespace-nowrap"
                            variant="outline"
                        >
                            {Icon && <Icon className="mr-1 size-3 shrink-0" />}
                            <span className="truncate">{label}</span>
                        </Badge>
                    );
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('settings.providers.type')}
                    />
                ),
                meta: { searchable: true },
                minSize: 110,
                size: 160,
            },
            {
                accessorKey: 'createdAt',
                cell: ({ row }) => {
                    const dateString = row.getValue('createdAt') as string;

                    return <div className="text-sm">{formatDate(new Date(dateString), dateLocale)}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('settings.providers.created')}
                    />
                ),
                meta: { columnMenuLabel: t('settings.providers.created') },
                size: 120,
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

                    return <div className="text-sm">{formatDate(new Date(dateString), dateLocale)}</div>;
                },
                header: ({ column }) => (
                    <DataTableColumnHeader
                        column={column}
                        title={t('settings.providers.updated')}
                    />
                ),
                size: 120,
                sortingFn: (rowA, rowB) => {
                    const dateA = new Date(rowA.getValue('updatedAt') as string);
                    const dateB = new Date(rowB.getValue('updatedAt') as string);

                    return dateA.getTime() - dateB.getTime();
                },
            },
            {
                cell: ({ row }) => {
                    const provider = row.original;

                    return (
                        <div className="flex justify-end opacity-0 transition-opacity group-hover:opacity-100">
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        aria-label={t('common.openMenu')}
                                        className="size-8 p-0"
                                        variant="ghost"
                                    >
                                        <Ellipsis />
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent
                                    align="end"
                                    className="min-w-24"
                                >
                                    <DropdownMenuItem onClick={() => handleProviderEdit(provider.id)}>
                                        <Pencil className="size-3" />
                                        {t('common.edit')}
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={() => handleProviderClone(provider.id)}>
                                        <Copy className="size-4" />
                                        {t('common.clone')}
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuItem
                                        disabled={isDeleteLoading && deletingProvider?.id === provider.id}
                                        onClick={() => handleProviderDeleteDialogOpen(provider)}
                                    >
                                        {isDeleteLoading && deletingProvider?.id === provider.id ? (
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
                meta: { preventRowClick: true },
                size: 48,
            },
        ],
        [
            dateLocale,
            deletingProvider,
            handleProviderClone,
            handleProviderDeleteDialogOpen,
            handleProviderEdit,
            isDeleteLoading,
            t,
        ],
    );

    const renderSubComponent = ({ row }: { row: Row<Provider> }) => {
        const provider = row.original;
        const { agents } = provider;

        if (!agents) {
            return (
                <div className="text-muted-foreground p-4 text-sm">{t('settings.providers.noAgentConfiguration')}</div>
            );
        }

        const formatValue = (value: boolean | number | string): string => {
            if (typeof value === 'boolean') {
                return value ? t('common.yes') : t('common.no');
            }

            if (value === 'high' || value === 'medium' || value === 'low') {
                return t(`common.${value}`);
            }

            return String(value);
        };

        const getFields = (obj: unknown, path: string[] = []): { label: string; value: string }[] => {
            if (!obj || typeof obj !== 'object') {
                return [];
            }

            return Object.entries(obj as Record<string, unknown>)
                .filter(([key, value]) => key !== '__typename' && !!value)
                .flatMap(([key, value]) => {
                    const fieldPath = [...path, key];

                    return typeof value === 'object'
                        ? getFields(value, fieldPath)
                        : [
                              {
                                  label: translateProviderFieldName(fieldPath.join('.'), t),
                                  value: formatValue(value as boolean | number | string),
                              },
                          ];
                });
        };

        const agentTypes = Object.entries(agents)
            .filter(([key]) => key !== '__typename')
            .map(([key, data]) => ({
                data,
                key,
                name: translateAgentName(key, t),
            }))
            .sort((a, b) => a.name.localeCompare(b.name));

        return (
            <div className="bg-muted/20 border-t p-4">
                <h4 className="font-medium">{t('settings.provider.agentConfigurations')}</h4>
                <hr className="border-muted-foreground/20 my-4" />
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 2xl:grid-cols-5">
                    {agentTypes.map(({ data, key, name }) => {
                        const fields = data ? getFields(data) : [];

                        return (
                            <div
                                className="flex flex-col gap-2"
                                key={key}
                            >
                                <div className="text-sm font-medium">{name}</div>
                                {fields.length > 0 ? (
                                    <div className="flex flex-col gap-1 text-sm">
                                        {fields.map(({ label, value }) => (
                                            <div key={label}>
                                                <span className="text-muted-foreground">{label}:</span> {value}
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-muted-foreground text-sm">
                                        {t('settings.providers.noConfiguration')}
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    };

    const renderRowContextMenu = useCallback(
        (provider: Provider) => (
            <>
                <ContextMenuItem onClick={() => handleProviderEdit(provider.id)}>
                    <Pencil />
                    {t('common.edit')}
                </ContextMenuItem>
                <ContextMenuItem onClick={() => handleProviderClone(provider.id)}>
                    <Copy />
                    {t('common.clone')}
                </ContextMenuItem>
                <ContextMenuSeparator />
                <ContextMenuItem
                    disabled={isDeleteLoading && deletingProvider?.id === provider.id}
                    onClick={() => handleProviderDeleteDialogOpen(provider)}
                >
                    <Trash />
                    {isDeleteLoading && deletingProvider?.id === provider.id
                        ? t('common.deleting')
                        : t('common.delete')}
                </ContextMenuItem>
            </>
        ),
        [deletingProvider, handleProviderClone, handleProviderDeleteDialogOpen, handleProviderEdit, isDeleteLoading, t],
    );

    if (isLoading) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsProvidersHeader />
                <StatusCard
                    description={t('settings.providers.loadingDescription')}
                    icon={<Loader2 className="text-muted-foreground size-16 animate-spin" />}
                    title={t('settings.providers.loadingTitle')}
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsProvidersHeader />
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>{t('settings.providers.loadingError')}</AlertTitle>
                    <AlertDescription>{error.message}</AlertDescription>
                </Alert>
            </div>
        );
    }

    const providers = data?.settingsProviders?.userDefined || [];

    if (providers.length === 0) {
        return (
            <div className="flex flex-col gap-4">
                <SettingsProvidersHeader />
                <StatusCard
                    action={
                        <Button
                            onClick={() => navigate('/settings/providers/new')}
                            variant="secondary"
                        >
                            <Plus className="size-4" />
                            {t('settings.providers.add')}
                        </Button>
                    }
                    description={t('settings.providers.emptyDescription')}
                    icon={<Settings className="text-muted-foreground size-8" />}
                    title={t('settings.providers.emptyTitle')}
                />
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-4">
            <SettingsProvidersHeader />

            {/* Delete Error Alert */}
            {(deleteError || deleteErrorMessage) && (
                <Alert variant="destructive">
                    <AlertCircle className="size-4" />
                    <AlertTitle>{t('settings.providers.deletingErrorTitle')}</AlertTitle>
                    <AlertDescription>{deleteError?.message || deleteErrorMessage}</AlertDescription>
                </Alert>
            )}

            <DataTable<Provider>
                columns={columns}
                data={providers}
                empty={{ entityName: 'providers' }}
                filterPlaceholder={t('settings.providers.filterPlaceholder')}
                filterValue={filter}
                onFilterChange={setFilter}
                onPageChange={handlePageChange}
                pageIndex={currentPage}
                renderRowContextMenu={renderRowContextMenu}
                renderSubComponent={renderSubComponent}
            />

            <ConfirmationDialog
                cancelText={t('common.cancel')}
                confirmText={t('common.delete')}
                description={t('settings.providers.deleteDescription', { name: deletingProvider?.name ?? '' })}
                handleConfirm={() => handleProviderDelete(deletingProvider?.id)}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={deletingProvider?.name}
                itemType="provider"
                title={t('settings.providers.deleteTitle')}
            />
        </div>
    );
}

function SettingsProvidersHeader() {
    const navigate = useNavigate();
    const { t } = useLocale();

    const handleProviderCreate = (providerType: string) => {
        navigate(`/settings/providers/new?type=${providerType}`);
    };

    return (
        <div className="flex items-center justify-between gap-4">
            <p className="text-muted-foreground min-w-0 flex-1 truncate">{t('settings.providers.manage')}</p>

            {/*
             * "Create Provider" is a dropdown trigger, not a submit-style action — it
             * opens a menu listing provider types (OpenAI, Anthropic, Custom, …). The
             * `<ChevronDown />` icon plus Radix's `aria-haspopup="menu"` already signal
             * "menu opens" to sighted and AT users; the explicit aria-label adds the
             * intent ("create provider") so screen readers don't just announce
             * "Create Provider, menu" but "Create provider, choose type, menu".
             */}
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button
                        aria-label={t('settings.providers.createAria')}
                        className="shrink-0"
                        variant="secondary"
                    >
                        {t('settings.createProvider')}
                        <ChevronDown className="size-4" />
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                    align="end"
                    style={{
                        width: 'var(--radix-dropdown-menu-trigger-width)',
                    }}
                >
                    {providerTypes.map(({ label, type }) => {
                        const Icon = providerIcons[type];
                        const displayLabel = type === ProviderType.Custom ? t('settings.providers.custom') : label;

                        return (
                            <DropdownMenuItem
                                key={type}
                                onClick={() => handleProviderCreate(type)}
                            >
                                {Icon && <Icon className="size-4" />}
                                {displayLabel}
                            </DropdownMenuItem>
                        );
                    })}
                </DropdownMenuContent>
            </DropdownMenu>
        </div>
    );
}

export default SettingsProviders;
