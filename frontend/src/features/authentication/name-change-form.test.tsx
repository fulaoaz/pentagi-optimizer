import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { patchUser, put, refreshAuthInfo } = vi.hoisted(() => ({
    patchUser: vi.fn(),
    put: vi.fn(),
    refreshAuthInfo: vi.fn().mockResolvedValue(undefined),
}));

// Keep the real error-mapping helper (`resolveApiErrorMessage`) — only the network call is stubbed.
vi.mock('@/lib/axios', async (importOriginal) => {
    const actual = await importOriginal<typeof import('@/lib/axios')>();

    return { ...actual, api: { ...actual.api, put } };
});

vi.mock('@/providers/user-provider', () => ({
    useUser: () => ({ authInfo: { user: { name: 'Old Name' } }, patchUser, refreshAuthInfo }),
}));

vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

import { LocaleProvider } from '@/providers/locale-provider';
import { NameChangeForm } from './name-change-form';

const apiError = (code: string, msg: string) => ({ response: { data: { code, msg, status: 'error' } } });

beforeEach(() => {
    put.mockReset().mockResolvedValue({ status: 'success' });
    refreshAuthInfo.mockClear();
    patchUser.mockClear();
});

describe('NameChangeForm', () => {
    it('seeds the field with the current name', () => {
        render(<LocaleProvider><NameChangeForm /></LocaleProvider>);

        expect((screen.getByLabelText('显示名称') as HTMLInputElement).value).toBe('Old Name');
    });

    it('submits the trimmed name and refreshes auth before closing', async () => {
        const user = userEvent.setup();
        const onSuccess = vi.fn();
        render(<LocaleProvider><NameChangeForm onSuccess={onSuccess} /></LocaleProvider>);

        const input = screen.getByLabelText('显示名称');
        await user.clear(input);
        await user.type(input, '  New Name  ');
        await user.click(screen.getByRole('button', { name: '更新名称' }));

        await waitFor(() => expect(onSuccess).toHaveBeenCalledOnce());
        expect(put).toHaveBeenCalledWith('/user/name', { name: 'New Name' });
        expect(patchUser).toHaveBeenCalledWith({ name: 'New Name' });
        expect(refreshAuthInfo).toHaveBeenCalledOnce();
    });

    it('blocks an empty name without calling the API', async () => {
        const user = userEvent.setup();
        render(<LocaleProvider><NameChangeForm /></LocaleProvider>);

        await user.clear(screen.getByLabelText('显示名称'));
        await user.click(screen.getByRole('button', { name: '更新名称' }));

        expect(await screen.findByText('请填写名称')).toBeInTheDocument();
        expect(put).not.toHaveBeenCalled();
    });

    it('maps the user-not-found code to friendly copy', async () => {
        const user = userEvent.setup();
        put.mockRejectedValueOnce(apiError('Users.NotFound', 'user not found'));
        render(<LocaleProvider><NameChangeForm /></LocaleProvider>);

        const input = screen.getByLabelText('显示名称');
        await user.clear(input);
        await user.type(input, 'Whoever');
        await user.click(screen.getByRole('button', { name: '更新名称' }));

        expect(await screen.findByText('User not found')).toBeInTheDocument();
    });
});
