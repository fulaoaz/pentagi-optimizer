import { ArrowLeft, FileText, Key, Plug, Settings as SettingsIcon } from 'lucide-react';
import { useMemo } from 'react';
import { NavLink, Outlet, useLocation, useParams } from 'react-router-dom';

import { Separator } from '@/components/ui/separator';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupContent,
    SidebarHeader,
    SidebarInset,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarProvider,
    SidebarTrigger,
} from '@/components/ui/sidebar';
import { useLocale } from '@/hooks/use-locale';

export interface MenuItem {
    icon?: React.ReactNode;
    id: string;
    isActive?: boolean;
    path: string;
    titleKey: string;
}

interface SettingsSidebarMenuItemProps {
    item: MenuItem;
}

const menuItems: readonly MenuItem[] = [
    {
        icon: <Plug className="size-4" />,
        id: 'providers',
        path: '/settings/providers',
        titleKey: 'nav.providers',
    },
    {
        icon: <FileText className="size-4" />,
        id: 'prompts',
        path: '/settings/prompts',
        titleKey: 'nav.prompts',
    },
    {
        icon: <Key className="size-4" />,
        id: 'api-tokens',
        path: '/settings/api-tokens',
        titleKey: 'nav.apiTokens',
    },
] as const;

function SettingsHeader() {
    const location = useLocation();
    const params = useParams();
    const { t } = useLocale();

    const title = useMemo(() => {
        const path = location.pathname;

        if (path === '/settings/providers/new') {
            return t('settings.createProvider');
        }

        if (path.startsWith('/settings/providers/') && params.providerId && params.providerId !== 'new') {
            return t('settings.editProvider');
        }

        if (path === '/settings/prompts/new') {
            return t('settings.createPrompt');
        }

        if (path.startsWith('/settings/prompts/') && params.promptId && params.promptId !== 'new') {
            return t('settings.editPrompt');
        }

        const activeItem = menuItems.find((item) => path.startsWith(item.path));

        return activeItem ? t(activeItem.titleKey) : t('nav.settings');
    }, [location.pathname, params, t]);

    return (
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator
                className="mr-2 h-4"
                orientation="vertical"
            />
            <h1 className="text-lg font-semibold">{title}</h1>
        </header>
    );
}

function SettingsLayout() {
    return (
        <SidebarProvider>
            <div className="flex h-screen w-full overflow-hidden">
                <SettingsSidebar />
                <SidebarInset className="flex flex-1 flex-col">
                    <SettingsHeader />
                    <main className="min-h-0 flex-1 overflow-auto p-4">
                        <Outlet />
                    </main>
                </SidebarInset>
            </div>
        </SidebarProvider>
    );
}

function SettingsSidebar() {
    const { t } = useLocale();

    return (
        <Sidebar collapsible="icon">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem className="flex items-center gap-2">
                        <div className="flex aspect-square size-8 items-center justify-center">
                            <SettingsIcon className="size-6" />
                        </div>
                        <div className="grid flex-1 text-left leading-tight">
                            <span className="truncate font-semibold">{t('nav.settings')}</span>
                        </div>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                <SidebarGroup>
                    <SidebarGroupContent>
                        <SidebarMenu>
                            {menuItems.map((item) => (
                                <SettingsSidebarMenuItem
                                    item={item}
                                    key={item.id}
                                />
                            ))}
                        </SidebarMenu>
                    </SidebarGroupContent>
                </SidebarGroup>
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenuButton asChild>
                    <NavLink to="/flows">
                        <ArrowLeft className="size-4" />
                        {t('settings.backToApp')}
                    </NavLink>
                </SidebarMenuButton>
            </SidebarFooter>
        </Sidebar>
    );
}

function SettingsSidebarMenuItem({ item }: SettingsSidebarMenuItemProps) {
    const location = useLocation();
    const isActive = location.pathname.startsWith(item.path);
    const { t } = useLocale();

    return (
        <SidebarMenuItem>
            <SidebarMenuButton
                asChild
                isActive={isActive}
            >
                <NavLink to={item.path}>
                    {item.icon}
                    {t(item.titleKey)}
                </NavLink>
            </SidebarMenuButton>
        </SidebarMenuItem>
    );
}

export default SettingsLayout;
