import { NetworkStatus } from '@apollo/client';
import { createContext, useCallback, useContext, useEffect, useMemo } from 'react';
import { toast } from 'sonner';

import type { FlowFormValues } from '@/features/flows/flow-form';
import type { FlowFragmentFragment, FlowsQuery } from '@/graphql/types';

import {
    useCreateAssistantMutation,
    useCreateFlowMutation,
    useDeleteFlowMutation,
    useFinishFlowMutation,
    useFlowCreatedSubscription,
    useFlowDeletedSubscription,
    useFlowsQuery,
    useFlowUpdatedSubscription,
} from '@/graphql/types';
import { useLocale } from '@/hooks/use-locale';
import { Log } from '@/lib/log';

export type Flow = FlowFragmentFragment;

interface FlowsContextValue {
    createFlow: (values: FlowFormValues) => Promise<null | string>;
    createFlowWithAssistant: (values: FlowFormValues) => Promise<null | string>;
    deleteFlow: (flow: Flow) => Promise<boolean>;
    finishFlow: (flow: Flow) => Promise<boolean>;
    flows: Array<Flow>;
    flowsData: FlowsQuery | undefined;
    flowsError: Error | undefined;
    isLoading: boolean;
}

const FlowsContext = createContext<FlowsContextValue | undefined>(undefined);

interface FlowsProviderProps {
    children: React.ReactNode;
}

export function FlowsProvider({ children }: FlowsProviderProps) {
    const { t } = useLocale();

    const {
        data: flowsData,
        error: flowsError,
        loading,
        networkStatus,
    } = useFlowsQuery({
        notifyOnNetworkStatusChange: true,
    });

    const isLoading = loading && networkStatus === NetworkStatus.loading;
    const flows = useMemo(() => flowsData?.flows ?? [], [flowsData?.flows]);

    useFlowCreatedSubscription();
    useFlowDeletedSubscription();
    useFlowUpdatedSubscription();

    useEffect(() => {
        if (flowsError) {
            toast.error(t('flow.provider.loadListFailed'), {
                description: flowsError.message,
            });
            Log.error('Error loading flows:', flowsError);
        }
    }, [flowsError, t]);

    const [createFlowMutation] = useCreateFlowMutation();
    const [createAssistantMutation] = useCreateAssistantMutation();
    const [deleteFlowMutation] = useDeleteFlowMutation();
    const [finishFlowMutation] = useFinishFlowMutation();

    const createFlow = useCallback(
        async (values: FlowFormValues) => {
            const { message, providerName, resourceIds } = values;

            const input = message.trim();
            const modelProvider = providerName.trim();

            if (!input || !modelProvider) {
                return null;
            }

            try {
                const { data } = await createFlowMutation({
                    variables: {
                        input,
                        modelProvider,
                        resourceIds: resourceIds?.length ? resourceIds : undefined,
                    },
                });

                if (data?.createFlow?.id) {
                    return data.createFlow.id;
                }

                return null;
            } catch (error) {
                const description = error instanceof Error ? error.message : t('flow.provider.createFlowError');
                toast.error(t('flow.provider.createFlowFailed'), {
                    description,
                });
                Log.error('Error creating flow:', error);

                return null;
            }
        },
        [createFlowMutation, t],
    );

    const createFlowWithAssistant = useCallback(
        async (values: FlowFormValues) => {
            const { message, providerName, resourceIds, useAgents } = values;

            const input = message.trim();
            const modelProvider = providerName.trim();

            if (!input || !modelProvider) {
                return null;
            }

            try {
                const { data } = await createAssistantMutation({
                    variables: {
                        flowId: '0',
                        input,
                        modelProvider,
                        resourceIds: resourceIds?.length ? resourceIds : undefined,
                        useAgents,
                    },
                });

                if (data?.createAssistant?.flow?.id) {
                    return data.createAssistant.flow.id;
                }

                return null;
            } catch (error) {
                const description = error instanceof Error ? error.message : t('flow.provider.createAssistantError');
                toast.error(t('flow.provider.createAssistantFailed'), {
                    description,
                });
                Log.error('Error creating assistant:', error);

                return null;
            }
        },
        [createAssistantMutation, t],
    );

    const deleteFlow = useCallback(
        async (flow: Flow) => {
            const { id: flowId, title } = flow;

            if (!flowId) {
                return false;
            }

            const flowDescription = t('flow.provider.description', {
                id: flowId,
                title: title || t('flow.provider.untitled'),
            });

            const loadingToastId = toast.loading(t('flow.provider.deleting'), {
                description: flowDescription,
            });

            try {
                await deleteFlowMutation({
                    variables: { flowId },
                });

                toast.success(t('flow.provider.deleted'), {
                    description: flowDescription,
                    id: loadingToastId,
                });

                return true;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : t('flow.provider.deleteFlowError');
                toast.error(errorMessage, {
                    description: flowDescription,
                    id: loadingToastId,
                });
                Log.error('Error deleting flow:', error);

                return false;
            }
        },
        [deleteFlowMutation, t],
    );

    const finishFlow = useCallback(
        async (flow: Flow) => {
            const { id: flowId, title } = flow;

            if (!flowId) {
                return false;
            }

            const flowDescription = t('flow.provider.description', {
                id: flowId,
                title: title || t('flow.provider.untitled'),
            });

            const loadingToastId = toast.loading(t('flow.provider.finishing'), {
                description: flowDescription,
            });

            try {
                await finishFlowMutation({
                    variables: { flowId },
                });

                toast.success(t('flow.provider.finished'), {
                    description: flowDescription,
                    id: loadingToastId,
                });

                return true;
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : t('flow.provider.finishFlowError');
                toast.error(errorMessage, {
                    description: flowDescription,
                    id: loadingToastId,
                });
                Log.error('Error finishing flow:', error);

                return false;
            }
        },
        [finishFlowMutation, t],
    );

    const value = useMemo(
        () => ({
            createFlow,
            createFlowWithAssistant,
            deleteFlow,
            finishFlow,
            flows,
            flowsData,
            flowsError,
            isLoading,
        }),
        [createFlow, createFlowWithAssistant, deleteFlow, finishFlow, flows, flowsData, flowsError, isLoading],
    );

    return <FlowsContext.Provider value={value}>{children}</FlowsContext.Provider>;
}

export function useFlows() {
    const context = useContext(FlowsContext);

    if (context === undefined) {
        throw new Error('useFlows must be used within FlowsProvider');
    }

    return context;
}
