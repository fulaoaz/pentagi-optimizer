import { render, screen } from '@testing-library/react';
import { ChevronDown, Plus } from 'lucide-react';
import { describe, expect, it } from 'vitest';

import { HeaderButton } from './header-button';

describe('HeaderButton', () => {
    it('keeps the action accessible while collapsing its label on mobile', () => {
        render(
            <HeaderButton
                endIcon={<ChevronDown />}
                icon={<Plus />}
                label="Create flow"
            />,
        );

        const button = screen.getByRole('button', { name: 'Create flow' });

        expect(button).toHaveClass('min-h-11', 'min-w-11', 'md:min-h-0', 'md:min-w-0', 'md:w-auto');
        expect(button).toHaveTextContent('Create flow');
        expect(button.querySelector('span.hidden')).toBeInTheDocument();
    });

    it('uses an explicit accessible label when the visible label is not text', () => {
        render(
            <HeaderButton
                aria-label="Open actions"
                icon={<Plus />}
                label={<span>Actions</span>}
            />,
        );

        expect(screen.getByRole('button', { name: 'Open actions' })).toBeInTheDocument();
    });
});
