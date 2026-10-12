import { BookOpen, Bot, Code2, Compass, Database, FileSearch, Search, Sparkles, Wrench } from 'lucide-react';
import { Link } from 'react-router-dom';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useLocale } from '@/hooks/use-locale';
import { routes } from '@/lib/routes';

const skillCatalog = [
    { icon: Compass, key: 'recon', tags: ['Searcher', 'Pentester'] },
    { icon: FileSearch, key: 'analysis', tags: ['Primary Agent', 'Adviser'] },
    { icon: Code2, key: 'coding', tags: ['Coder', 'Generator'] },
    { icon: Wrench, key: 'execution', tags: ['Installer', 'Tooling'] },
    { icon: Database, key: 'knowledge', tags: ['Memorist', 'Enricher'] },
    { icon: Bot, key: 'orchestration', tags: ['Assistant', 'Reflector'] },
] as const;

function SkillsCenter() {
    const { t } = useLocale();

    return (
        <div className="flex flex-col gap-6 p-4">
            <header className="flex flex-wrap items-start justify-between gap-4">
                <div>
                    <h1 className="flex items-center gap-2 text-2xl font-semibold">
                        <Sparkles className="size-6 text-primary" />
                        {t('integrations.skills.title')}
                    </h1>
                    <p className="text-muted-foreground mt-2 max-w-3xl">{t('integrations.skills.description')}</p>
                </div>
                <Button
                    asChild
                    variant="outline"
                >
                    <Link to={routes.settings.prompts}>{t('integrations.skills.openPrompts')}</Link>
                </Button>
            </header>

            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {skillCatalog.map(({ icon: Icon, key, tags }) => (
                    <Card
                        className="group transition-colors hover:border-primary/60"
                        key={key}
                    >
                        <CardHeader>
                            <div className="flex items-start justify-between gap-3">
                                <div className="flex items-center gap-3">
                                    <div className="bg-primary/10 text-primary rounded-lg p-2">
                                        <Icon className="size-5" />
                                    </div>
                                    <CardTitle className="text-base">{t(`integrations.skills.${key}.title`)}</CardTitle>
                                </div>
                                <Badge variant="secondary">{t('integrations.skills.builtIn')}</Badge>
                            </div>
                            <CardDescription>{t(`integrations.skills.${key}.description`)}</CardDescription>
                        </CardHeader>
                        <CardContent className="flex flex-wrap gap-2">
                            {tags.map((tag) => (
                                <Badge
                                    key={tag}
                                    variant="outline"
                                >
                                    {tag}
                                </Badge>
                            ))}
                        </CardContent>
                    </Card>
                ))}
            </div>

            <div className="grid gap-4 md:grid-cols-2">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <BookOpen className="size-5" />
                            {t('integrations.skills.promptTitle')}
                        </CardTitle>
                        <CardDescription>{t('integrations.skills.promptDescription')}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button
                            asChild
                            size="sm"
                        >
                            <Link to={routes.settings.prompts}>{t('integrations.skills.managePrompts')}</Link>
                        </Button>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-base">
                            <Search className="size-5" />
                            {t('integrations.skills.providerTitle')}
                        </CardTitle>
                        <CardDescription>{t('integrations.skills.providerDescription')}</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Button
                            asChild
                            size="sm"
                            variant="outline"
                        >
                            <Link to={routes.settings.providers}>{t('integrations.skills.manageProviders')}</Link>
                        </Button>
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

export default SkillsCenter;
