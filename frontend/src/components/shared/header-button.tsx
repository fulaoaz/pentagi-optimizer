import type { ReactNode } from 'react';

import { Button, type ButtonProps } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface HeaderButtonProps extends Omit<ButtonProps, 'children'> {
    endIcon?: ReactNode;
    icon: ReactNode;
    label: ReactNode;
}

export function HeaderButton({
    'aria-label': ariaLabel,
    className,
    endIcon,
    icon,
    label,
    size = 'sm',
    ...props
}: HeaderButtonProps) {
    const accessibleLabel = ariaLabel ?? (typeof label === 'string' ? label : undefined);

    return (
        <Button
            aria-label={accessibleLabel}
            className={cn('min-h-11 min-w-11 px-0 md:min-h-0 md:w-auto md:min-w-0 md:px-3', className)}
            size={size}
            {...props}
        >
            {icon}
            <span className="hidden md:inline">{label}</span>
            {endIcon ? <span className="hidden md:inline-flex">{endIcon}</span> : null}
        </Button>
    );
}
