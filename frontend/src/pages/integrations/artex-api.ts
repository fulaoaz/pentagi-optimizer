import { z } from 'zod';

import { api, unwrapApiResponse } from '@/lib/axios';

const id = z.union([z.string().regex(/^[1-9][0-9]*$/), z.number().int().positive()]).transform(String);
const text = z
    .string()
    .nullish()
    .transform((value) => value ?? '');

const taskSchema = z.object({
    description: text,
    goal: text,
    id,
    name: text,
    status: z.string(),
});
const assetSchema = z.object({
    app_name: text,
    domain: text,
    id,
    ip: text,
    port: z.number().nullish(),
    service_name: text,
    type: z.string(),
    url: text,
});
const findingSchema = z.object({
    evidence: text,
    id,
    name: text,
    report: text,
    severity: text,
    status: text,
    summary: text,
    task_id: id.nullish(),
    vulnclass: text,
});
const retestSchema = z.object({
    evidence: text,
    id,
    notes: text,
    status: text,
    summary: text,
    verdict: text,
});
const statusSchema = z.object({
    canControl: z.boolean(),
    canCreate: z.boolean(),
    canShareModels: z.boolean().optional().default(false),
    enabled: z.boolean(),
    errorCode: z.string().optional(),
    healthy: z.boolean(),
    modelSharingAvailable: z.boolean().optional().default(false),
    version: z.string().optional(),
});
const tasksSchema = z.object({ tasks: z.array(taskSchema).nullable() });
const assetsSchema = z.object({ assets: z.array(assetSchema).nullable(), total: z.number().nonnegative() });
const findingsSchema = z.object({ items: z.array(findingSchema).nullable(), total: z.number().nonnegative() });
const activitySchema = z.object({
    cursor: z.number().int().nonnegative(),
    items: z.array(z.object({ kind: text, seq: z.number(), summary: text, ts: text, worker: text })).nullable(),
});
const retestsSchema = z.object({ retests: z.array(retestSchema).nullable() });
const unifiedRunSchema = z.object({
    bindingId: z.string().optional(),
    bindingRevision: z.string().optional(),
    createdAt: z.string(),
    detailPath: z.string(),
    engine: z.enum(['pentagi', 'artex']),
    engineTaskId: z.string().optional(),
    id: z.string(),
    model: z.string().optional(),
    rawStatus: z.string().optional(),
    status: z.string(),
    submitState: z.string().optional(),
    title: z.string(),
    updatedAt: z.string(),
});
const unifiedRunsSchema = z.object({ items: z.array(unifiedRunSchema), total: z.number().nonnegative() });
const unifiedReportSchema = z.object({
    artifacts: z.array(
        z.object({
            capturedAt: z.string().optional(),
            content: z.unknown().optional(),
            id: z.string(),
            kind: z.string(),
            representation: z.string().optional(),
        }),
    ),
    engine: z.enum(['pentagi', 'artex']),
    engineTaskId: z.string().optional(),
    generatedAt: z.string(),
    id: z.string(),
    model: z.string().optional(),
    report: z.string().optional(),
    reportFormat: z.string().optional(),
    status: z.string(),
    title: z.string(),
});
const evidenceSnapshotSchema = z.object({
    capturedAt: z.string(),
    content: z.unknown().optional(),
    contentSha256: z.string(),
    engine: z.string(),
    id: z.string().uuid(),
    kind: z.string(),
    parentRecordId: z.string().optional(),
    redacted: z.boolean(),
    remoteRecordId: z.string(),
    remoteTaskId: z.string(),
    representation: z.string(),
    runId: z.string().uuid(),
    sequence: z.union([z.string(), z.number()]),
    version: z.number().int().positive(),
});
const evidenceSnapshotPageSchema = z.object({
    hasMore: z.boolean(),
    items: z.array(evidenceSnapshotSchema),
    nextCursor: z.string().optional(),
});

const providerChoiceSchema = z.object({
    compatible: z.boolean(),
    defaultModel: z.string(),
    models: z.array(z.string()),
    name: z.string(),
    reason: z.string().optional(),
    ref: z.string(),
    type: z.string(),
});
const bindingSchema = z.object({
    id: z.string().uuid(),
    providerRef: z.string(),
    remoteProfileId: z.number(),
    revision: z.string(),
    spec: z.object({
        base_url: z.string(),
        max_tokens: z.number(),
        model: z.string(),
        reasoning_effort: z.string().optional(),
    }),
    state: z.string(),
    updatedAt: z.string(),
});
export type ArtexAsset = z.infer<typeof assetSchema>;
export type ArtexBinding = z.infer<typeof bindingSchema>;

export type ArtexFinding = z.infer<typeof findingSchema>;
export type ArtexEvidenceSnapshot = z.infer<typeof evidenceSnapshotSchema>;
export interface ArtexModelSelection {
    maxTokens: number;
    model: string;
    providerRef: string;
    reasoningEffort: string;
}
export type ArtexRetest = z.infer<typeof retestSchema>;
export type ArtexStatus = z.infer<typeof statusSchema>;
export type ArtexTask = z.infer<typeof taskSchema>;
export type UnifiedReport = z.infer<typeof unifiedReportSchema>;
export type UnifiedRun = z.infer<typeof unifiedRunSchema>;

async function read<T>(
    path: string,
    schema: z.ZodType<T>,
    signal?: AbortSignal,
    params?: Record<string, number | string>,
) {
    const result = unwrapApiResponse(await api.get<unknown>('/artex' + path, { params, signal, timeout: 130_000 }));
    const parsed = schema.safeParse(result);

    if (!parsed.success) {
        throw new Error('Artex.Response');
    }

    return parsed.data;
}

export const artexApi = {
    activity: (task: string, since: number, signal?: AbortSignal) =>
        read('/activity', activitySchema, signal, { since, task }),
    assets: (task: string, page: number, signal?: AbortSignal) => read('/assets', assetsSchema, signal, { page, task }),
    control: async (task: string, action: 'pause' | 'resume') =>
        unwrapApiResponse(
            await api.post('/artex/tasks/' + encodeURIComponent(task) + '/control', { action }, { timeout: 130_000 }),
        ),
    create: async (
        description: string,
        goal: string,
        timeoutSeconds: number,
        bindingId: string,
        idempotencyKey = crypto.randomUUID(),
    ) => {
        const result = unwrapApiResponse(
            await api.post(
                '/artex/tasks',
                { bindingId, description, goal, timeout_seconds: timeoutSeconds },
                { headers: { 'Idempotency-Key': idempotencyKey }, timeout: 130_000 },
            ),
        );

        return taskSchema.parse(result);
    },
    finding: (finding: string, signal?: AbortSignal) =>
        read('/findings/' + encodeURIComponent(finding), findingSchema, signal),
    findings: (task: string, page: number, signal?: AbortSignal) =>
        read('/findings', findingsSchema, signal, { page, task }),
    snapshot: (task: string, snapshotID: string, signal?: AbortSignal) =>
        read(
            '/tasks/' + encodeURIComponent(task) + '/snapshots/' + encodeURIComponent(snapshotID),
            evidenceSnapshotSchema,
            signal,
        ),
    snapshots: (task: string, before = '', signal?: AbortSignal) =>
        read('/tasks/' + encodeURIComponent(task) + '/snapshots', evidenceSnapshotPageSchema, signal, {
            before,
            limit: 50,
        }),
    modelBindings: (signal?: AbortSignal) => read('/model-bindings', z.array(bindingSchema), signal),
    modelProviders: (signal?: AbortSignal) => read('/model-providers', z.array(providerChoiceSchema), signal),
    retest: async (finding: string, notes: string) =>
        unwrapApiResponse(
            await api.post(
                '/artex/findings/' + encodeURIComponent(finding) + '/retest',
                { notes },
                { timeout: 130_000 },
            ),
        ),
    retests: (finding: string, signal?: AbortSignal) =>
        read('/findings/' + encodeURIComponent(finding) + '/retests', retestsSchema, signal),
    rotateModelKey: async (bindingId: string) =>
        bindingSchema.parse(
            unwrapApiResponse(
                await api.post(
                    '/artex/model-bindings/' + encodeURIComponent(bindingId) + '/rotate-key',
                    {},
                    { timeout: 130_000 },
                ),
            ),
        ),
    status: (signal?: AbortSignal) => read('/status', statusSchema, signal),
    syncModel: async (input: ArtexModelSelection) =>
        bindingSchema.parse(unwrapApiResponse(await api.post('/artex/model-bindings', input, { timeout: 130_000 }))),
    tasks: (signal?: AbortSignal) => read('/tasks', tasksSchema, signal),
};

export const executionApi = {
    list: async (signal?: AbortSignal) => {
        const result = unwrapApiResponse(await api.get<unknown>('/execution-runs', { signal, timeout: 130_000 }));
        const parsed = unifiedRunsSchema.safeParse(result);
        if (!parsed.success) throw new Error('Artex.Response');
        return parsed.data;
    },
    report: async (runID: string) => {
        const result = unwrapApiResponse(
            await api.get<unknown>('/execution-runs/' + encodeURIComponent(runID) + '/report', { timeout: 130_000 }),
        );
        const parsed = unifiedReportSchema.safeParse(result);
        if (!parsed.success) throw new Error('Artex.Response');
        return parsed.data;
    },
    reconcile: async (runID: string) => {
        const result = unwrapApiResponse(
            await api.post('/artex/runs/' + encodeURIComponent(runID) + '/reconcile', {}, { timeout: 130_000 }),
        );
        return result;
    },
};
