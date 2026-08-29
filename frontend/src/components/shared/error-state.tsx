import { AlertCircle, RefreshCw } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { useLocale } from '@/hooks/use-locale';
import { Empty, EmptyContent, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';

interface ErrorStateProps {
    message?: null | string;
    onRetry?: () => unknown;
    title: string;
}

export function ErrorState({ message, onRetry, title }: ErrorStateProps) {
    const { t } = useLocale();

    return (
        <Empty role="alert">
            <EmptyHeader>
                <EmptyMedia>
                    <AlertCircle className="text-destructive size-12" />
                </EmptyMedia>
                <EmptyTitle>{title}</EmptyTitle>
                {message ? <EmptyDescription>{message}</EmptyDescription> : null}
            </EmptyHeader>
            {onRetry ? (
                <EmptyContent>
                    <Button
                        onClick={() => onRetry()}
                        variant="secondary"
                    >
                        <RefreshCw />
                        {t('common.tryAgain')}
                    </Button>
                </EmptyContent>
            ) : null}
        </Empty>
    );
}
