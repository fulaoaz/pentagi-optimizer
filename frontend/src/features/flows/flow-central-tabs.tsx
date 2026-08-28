import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import FlowDashboard from '@/features/flows/dashboard/flow-dashboard';
import FlowAssistantMessages from '@/features/flows/messages/flow-assistant-messages';
import FlowAutomationMessages from '@/features/flows/messages/flow-automation-messages';
import { useFlowTabDetection } from '@/hooks/use-flow-tab-detection';
import { useLocale } from '@/hooks/use-locale';

function FlowCentralTabs() {
    const { t } = useLocale();
    const { handleTabChange, resolvedTab } = useFlowTabDetection();

    return (
        <Tabs
            className="flex size-full flex-col"
            onValueChange={handleTabChange}
            value={resolvedTab}
        >
            <div className="max-w-full">
                <ScrollArea className="w-full pb-3">
                    <TabsList className="flex w-fit">
                        <TabsTrigger value="automation">{t('flow.tabs.automation')}</TabsTrigger>
                        <TabsTrigger value="assistant">{t('flow.tabs.assistant')}</TabsTrigger>
                        <TabsTrigger value="dashboard">{t('flow.tabs.dashboard')}</TabsTrigger>
                    </TabsList>
                    <ScrollBar orientation="horizontal" />
                </ScrollArea>
            </div>

            <TabsContent
                className="mt-1 flex-1 overflow-auto pr-4"
                value="automation"
            >
                <FlowAutomationMessages />
            </TabsContent>
            <TabsContent
                className="mt-1 flex-1 overflow-auto pr-4"
                value="assistant"
            >
                <FlowAssistantMessages />
            </TabsContent>
            <TabsContent
                className="mt-1 flex-1 overflow-auto pr-4"
                value="dashboard"
            >
                <FlowDashboard />
            </TabsContent>
        </Tabs>
    );
}

export default FlowCentralTabs;
