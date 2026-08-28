import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { FlowDashboardOverview } from '@/features/flows/dashboard/flow-dashboard-overview';
import { useLocale } from '@/hooks/use-locale';
import { useFlow } from '@/providers/flow-provider';

function FlowDashboard() {
    const { t } = useLocale();
    const { flowId } = useFlow();

    if (!flowId) {
        return (
            <div className="text-muted-foreground flex items-center justify-center py-12">
                {t('flow.dashboard.selectFlow')}
            </div>
        );
    }

    return (
        <Tabs defaultValue="overview">
            <TabsList className="hidden">
                <TabsTrigger value="overview">{t('flow.dashboard.overview')}</TabsTrigger>
            </TabsList>

            <TabsContent value="overview">
                <FlowDashboardOverview flowId={flowId} />
            </TabsContent>
        </Tabs>
    );
}

export default FlowDashboard;
