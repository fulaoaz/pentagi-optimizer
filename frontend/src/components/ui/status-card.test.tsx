import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { StatusCard } from './status-card';

describe('StatusCard', () => {
    it('renders a consistent title, description, icon, and action surface', () => {
        render(
            <StatusCard
                action={<button type="button">Create one</button>}
                className="custom-card"
                description="Start by creating your first item."
                icon={<span aria-hidden="true">+</span>}
                title="Nothing here yet"
            />,
        );

        const card = screen.getByRole('heading', { name: 'Nothing here yet' }).closest('[data-slot="status-card"]');

        expect(card).toBeInTheDocument();
        expect(card).toHaveClass('custom-card');
        expect(screen.getByText('Start by creating your first item.')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Create one' })).toBeInTheDocument();
    });
});
