import { act, cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn(), privileges: ['flows.admin'] }));

vi.mock('@/providers/user-provider', () => ({ useUser: () => ({ authInfo: { privileges: mocks.privileges } }) }));
vi.mock('@/lib/axios', () => ({
    api: { get: mocks.get, post: mocks.post },
    getApiErrorMessage: (error: unknown, fallback: string) => (error instanceof Error ? error.message : fallback),
    unwrapApiResponse: (response: { data: unknown; status: string }) => {
        if (response.status !== 'success') {
            throw new Error('rejected');
        }

        return response.data;
    },
}));

import ArtexWorkbench from './artex-workbench';

const ok = (data: unknown) => ({ data, status: 'success' });
const connected = {
    canControl: true,
    canCreate: true,
    canShareModels: true,
    enabled: true,
    healthy: true,
    modelSharingAvailable: true,
    version: 'contract-test',
};
const sharedBinding = {
    id: '11111111-1111-4111-8111-111111111111',
    providerRef: 'saved:7',
    remoteProfileId: 51,
    revision: 'fixture-revision',
    spec: { base_url: 'https://model.invalid/v1', max_tokens: 0, model: 'shared-model' },
    state: 'ready',
    updatedAt: '2026-09-24T12:00:00Z',
};
const tasks = [
    { description: 'Alpha lab', goal: 'inventory Alpha', id: '1', status: 'running' },
    { description: 'Beta lab', goal: 'inventory Beta', id: '2', status: 'paused' },
];

function renderWorkbench() {
    return render(
        <MemoryRouter>
            <ArtexWorkbench />
        </MemoryRouter>,
    );
}

beforeEach(() => {
    mocks.get.mockReset();
    mocks.post.mockReset();
    mocks.privileges = ['flows.admin'];
    mocks.get.mockImplementation(async (path: string, config?: { params?: { task: string } }) => {
        if (path === '/artex/status') {
            return ok(connected);
        }

        if (path === '/artex/tasks') {
            return ok({ tasks });
        }

        if (path === '/artex/model-bindings') {
            return ok([sharedBinding]);
        }

        if (path === '/artex/model-providers') {
            return ok([
                {
                    compatible: true,
                    defaultModel: 'shared-model',
                    models: ['shared-model'],
                    name: 'Existing PentAGI provider',
                    ref: 'saved:7',
                    type: 'custom',
                },
            ]);
        }

        if (path === '/artex/findings') {
            return ok({
                items: [
                    {
                        id: '123',
                        name: config?.params?.task === '2' ? 'Beta finding' : 'Alpha finding',
                        status: 'open',
                    },
                ],
                total: 1,
            });
        }

        if (path === '/artex/findings/123') {
            return ok({
                evidence: '<script>untrusted</script>',
                id: '123',
                name: 'Finding detail',
                summary: 'Reproduction evidence',
            });
        }

        if (path === '/artex/findings/123/retests') {
            return ok({ retests: [] });
        }

        if (path === '/artex/assets') {
            return ok({ assets: [{ domain: 'lab.test', id: 5, type: 'subdomain' }], total: 1 });
        }

        throw new Error('Unexpected path: ' + path);
    });
    mocks.post.mockResolvedValue(ok({ created: true }));
});
afterEach(cleanup);

describe('ARTEX workbench', () => {
    it('synchronizes an existing provider without asking the browser for its key', async () => {
        mocks.post.mockResolvedValue(ok(sharedBinding));
        renderWorkbench();
        await screen.findByText('Alpha finding');
        fireEvent.click(screen.getByRole('button', { name: 'Shared models' }));
        const picker = await screen.findByLabelText('Model provider');
        fireEvent.change(picker, { target: { value: 'saved:7' } });
        expect(screen.getByLabelText('Model identifier')).toHaveValue('shared-model');
        expect(screen.queryByLabelText(/API key/i)).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'Synchronize model' }));
        await waitFor(() =>
            expect(mocks.post).toHaveBeenCalledWith(
                '/artex/model-bindings',
                {
                    maxTokens: 0,
                    model: 'shared-model',
                    providerRef: 'saved:7',
                    reasoningEffort: '',
                },
                expect.anything(),
            ),
        );
        await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
        await waitFor(() => expect(screen.getByRole('button', { name: 'Create task' })).toBeEnabled());
        fireEvent.click(screen.getByRole('button', { name: 'Create task' }));
        expect(screen.getByLabelText('Shared model version')).toHaveValue(sharedBinding.id);
    });

    it('does not submit a task without an explicit synchronized binding', async () => {
        renderWorkbench();
        await screen.findByText('Alpha finding');
        fireEvent.click(screen.getByRole('button', { name: 'Create task' }));
        fireEvent.change(screen.getByLabelText('Task description'), { target: { value: 'local test' } });
        fireEvent.change(screen.getByLabelText('Target, scope and objective'), { target: { value: 'local only' } });
        fireEvent.submit(screen.getByLabelText('Task description').closest('form')!);
        expect(mocks.post).not.toHaveBeenCalled();
    });
    it('pages activity with its cursor and preserves that page on refresh', async () => {
        const normal = mocks.get.getMockImplementation()!;
        mocks.get.mockImplementation((path: string, config?: { params?: { since?: number } }) => {
            if (path === '/artex/activity') {
                const since = config?.params?.since ?? 0;
                const items =
                    since === 0
                        ? Array.from({ length: 100 }, (_, index) => ({
                              seq: index + 1,
                              summary: 'Event ' + (index + 1),
                          }))
                        : [{ seq: 101, summary: 'Next activity page' }];

                return Promise.resolve(ok({ cursor: since === 0 ? 100 : 101, items }));
            }

            return normal(path, config);
        });
        renderWorkbench();
        await screen.findByText('Alpha finding');
        fireEvent.mouseDown(screen.getByRole('tab', { name: 'Activity' }), { button: 0, ctrlKey: false });
        await screen.findByText('Event 100');
        fireEvent.click(screen.getByRole('button', { name: 'Next page' }));
        await screen.findByText('Next activity page');
        fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
        await screen.findByText('Next activity page');
        const activityCalls = mocks.get.mock.calls.filter(([path]) => path === '/artex/activity');
        expect(activityCalls.map(([, options]) => options.params.since)).toEqual([0, 100, 100]);
    });
    it('does not request upstream data without administrator privileges', () => {
        mocks.privileges = ['settings.view', 'flows.create'];
        renderWorkbench();
        expect(screen.getByText(/Flow administrator access is required/)).toBeInTheDocument();
        expect(mocks.get).not.toHaveBeenCalled();
    });

    it('shows setup without querying disabled endpoints', async () => {
        mocks.get.mockResolvedValue(
            ok({ ...connected, canControl: false, canCreate: false, enabled: false, healthy: false }),
        );
        renderWorkbench();
        expect(await screen.findByText(/Configure the ARTEX connection/)).toBeInTheDocument();
        expect(mocks.get).toHaveBeenCalledTimes(1);
        expect(screen.getByRole('button', { name: 'Create task' })).toBeDisabled();
    });

    it('keeps the second task selected and ignores the first task response after switching', async () => {
        const normal = mocks.get.getMockImplementation()!;
        let resolveOld: (value: unknown) => void = () => undefined;
        mocks.get.mockImplementation((path: string, config?: { params?: { task: string } }) => {
            if (path === '/artex/findings' && config?.params?.task === '1') {
                return new Promise((resolve) => {
                    resolveOld = resolve;
                });
            }

            return normal(path, config);
        });
        renderWorkbench();
        fireEvent.click(await screen.findByRole('button', { name: /Beta lab/ }));
        expect(await screen.findByText('Beta finding')).toBeInTheDocument();
        await act(async () => resolveOld(ok({ items: [{ id: '99', name: 'Alpha stale' }], total: 1 })));
        expect(screen.queryByText('Alpha stale')).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Beta lab/ })).toHaveAttribute('aria-pressed', 'true');
        expect(mocks.get.mock.calls.filter(([path]) => path === '/artex/tasks')).toHaveLength(1);
        fireEvent.click(screen.getByRole('button', { name: 'Refresh' }));
        await waitFor(() =>
            expect(screen.getByRole('button', { name: /Beta lab/ })).toHaveAttribute('aria-pressed', 'true'),
        );
    });

    it('disables mutations in read-only mode', async () => {
        const normal = mocks.get.getMockImplementation()!;
        mocks.get.mockImplementation((path: string, config: unknown) =>
            path === '/artex/status'
                ? Promise.resolve(ok({ ...connected, canControl: false, canCreate: false }))
                : normal(path, config),
        );
        renderWorkbench();
        await screen.findByText('Alpha finding');
        expect(screen.getByRole('button', { name: 'Create task' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Pause' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Resume' })).toBeDisabled();
    });

    it('reads stable finding IDs, renders evidence as text and sends retest notes', async () => {
        renderWorkbench();
        fireEvent.click(await screen.findByText('Alpha finding'));
        expect(await screen.findByText('<script>untrusted</script>')).toBeInTheDocument();
        expect(document.querySelector('script')).toBeNull();
        fireEvent.change(screen.getByLabelText('Retest notes (optional)'), { target: { value: 'fixed local build' } });
        fireEvent.click(screen.getByRole('button', { name: 'Start retest' }));
        await waitFor(() =>
            expect(mocks.post).toHaveBeenCalledWith(
                '/artex/findings/123/retest',
                { notes: 'fixed local build' },
                expect.anything(),
            ),
        );
    });

    it('retains form input when task creation is rejected', async () => {
        mocks.post.mockResolvedValue({ status: 'error' });
        renderWorkbench();
        await screen.findByText('Alpha finding');
        fireEvent.click(screen.getByRole('button', { name: 'Create task' }));
        fireEvent.change(screen.getByLabelText('Shared model version'), { target: { value: sharedBinding.id } });
        fireEvent.change(screen.getByLabelText('Task description'), { target: { value: 'local test' } });
        fireEvent.change(screen.getByLabelText('Target, scope and objective'), { target: { value: 'lab.test only' } });
        fireEvent.submit(screen.getByLabelText('Task description').closest('form')!);
        await waitFor(() =>
            expect(mocks.post).toHaveBeenCalledWith(
                '/artex/tasks',
                {
                    bindingId: sharedBinding.id,
                    description: 'local test',
                    goal: 'lab.test only',
                    timeout_seconds: 1800,
                },
                expect.anything(),
            ),
        );
        await waitFor(() => expect(screen.getByLabelText('Task description')).toHaveValue('local test'));
        expect(screen.getByRole('dialog')).toBeInTheDocument();
    });
});
