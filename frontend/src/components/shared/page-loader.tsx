import { useLocale } from '@/hooks/use-locale';

function PageLoader() {
    const { t } = useLocale();

    return (
        <div className="grid h-screen w-full place-items-center">
            <p>{t('common.loading')}</p>
        </div>
    );
}

export default PageLoader;
