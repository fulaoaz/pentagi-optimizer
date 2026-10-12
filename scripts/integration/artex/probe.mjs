// Runs against real ARTEX/PostgreSQL in the isolated lab. The LLM is explicitly
// a protocol fixture; this does not claim real-model or vulnerability accuracy.
import { execFileSync } from 'node:child_process';
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';

const base = 'http://127.0.0.1:18787';
const fixture = 'http://127.0.0.1:18788';
const required = (name) => {
    if (!process.env[name]) throw new Error('Missing environment variable: ' + name);
    return process.env[name];
};
const password = required('ARTEX_LAB_ADMIN_PASSWORD');
const keys = [required('ARTEX_LAB_KEY_A'), required('ARTEX_LAB_KEY_B'), required('ARTEX_LAB_KEY_ROTATED')];
const reportPath = required('ARTEX_LAB_REPORT');
const statePath = required('ARTEX_LAB_STATE');
const mode = process.argv[2] ?? 'bootstrap';
const report = {
    observedAt: new Date().toISOString(), sourceCommit: 'bf7f414477832b77d2152539c0723dc691086522', mode,
    realComponents: ['ARTEX backend', 'PostgreSQL 16'], simulatedComponents: ['OpenAI protocol model fixture'],
    checks: [], outcome: 'running',
};
let token;
const pause = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const digest = (key) => createHash('sha256').update('Bearer ' + key).digest('hex');
const check = (name, condition, details = {}) => {
    report.checks.push({ name, passed: Boolean(condition), ...details });
    if (!condition) throw new Error('Check failed: ' + name);
    console.log('PASS ' + name);
};
async function api(method, path, body, authenticated = true) {
    const response = await fetch(base + path, {
        method, headers: { 'Content-Type': 'application/json', ...(authenticated && token ? { Authorization: 'Bearer ' + token } : {}) },
        ...(body === undefined ? {} : { body: JSON.stringify(body) }), signal: AbortSignal.timeout(15_000),
    });
    if (!response.ok) throw new Error(method + ' ' + path + ' HTTP ' + response.status);
    return response.json();
}
async function fixtureAPI(path, body) {
    const response = await fetch(fixture + path, { method: body ? 'POST' : 'GET',
        headers: { 'Content-Type': 'application/json' }, ...(body ? { body: JSON.stringify(body) } : {}), signal: AbortSignal.timeout(3_000) });
    if (!response.ok) throw new Error('Fixture request failed');
    return response.json();
}
async function until(name, fn, timeout = 15_000) {
    const end = Date.now() + timeout;
    while (Date.now() < end) {
        const result = await fn();
        if (result) return result;
        await pause(200);
    }
    throw new Error('Timeout: ' + name);
}
async function pauseTask(id) {
    const task = await api('GET', '/api/tasks/' + id);
    if (['running', 'created', 'queued'].includes(task.status)) await api('POST', '/api/tasks/' + id + '/control', { action: 'pause' });
    return api('GET', '/api/tasks/' + id);
}
async function profile(name, model, key, oldID) {
    const existing = (await api('GET', '/api/llm/profiles')).profiles.find((item) => item.name === name);
    const id = oldID ?? (existing ? Number(existing.id) : undefined);
    return api('POST', '/api/llm/profiles', {
        ...(id ? { id } : {}), name, model, format: 'openai', base_url: 'http://127.0.0.1:18999/v1',
        api_key: key, streaming: true, pool_exclude: true, max_tokens: 128, context_window_k: 32,
    });
}

try {
    const project = execFileSync('docker', ['inspect', '--format', '{{index .Config.Labels "com.docker.compose.project"}}', 'pentagi-artex-contract-lab-artex-1'], { encoding: 'utf8', windowsHide: true }).trim();
    check('owned_isolated_lab', project === 'pentagi-artex-contract-lab');
    const net = JSON.parse(execFileSync('docker', ['network', 'inspect', 'pentagi-artex-contract-lab_default'], { encoding: 'utf8', windowsHide: true }))[0];
    check('no_external_container_egress', net.Internal === true);
    const health = await api('GET', '/api/health', undefined, false);
    check('real_artex_health', health.ok && health.service === 'artex');
    const missingAuth = await fetch(base + '/api/tasks', { signal: AbortSignal.timeout(3_000) });
    check('tasks_require_authentication', missingAuth.status === 401);
    const authStatus = await api('GET', '/api/auth/status', undefined, false);
    if (!authStatus.initialized) await api('POST', '/api/auth/init', { password }, false);
    token = (await api('POST', '/api/auth/login', { username: 'ARTEX', password }, false)).token;
    check('login_token_issued', typeof token === 'string' && token.length > 20);

    if (mode === 'verify-restart') {
        const state = JSON.parse(readFileSync(statePath, 'utf8'));
        const task = await api('GET', '/api/tasks/' + state.taskID);
        const profiles = (await api('GET', '/api/llm/profiles')).profiles;
        check('paused_task_survives_restart', task.status === 'paused');
        check('profile_mapping_survives_restart', profiles.some((item) => Number(item.id) === state.profileB && item.model === 'lab-bound-b'));
        check('explicit_task_binding_survives_restart', task.llm_profile_ids.map(Number).includes(state.profileB));
        await fixtureAPI('/mode', { holdMs: 8_000 });
        await fixtureAPI('/reset', {});
        await api('POST', '/api/tasks/' + state.taskID + '/control', { action: 'resume' });
        await until('restored profile calls model', async () => (await fixtureAPI('/diagnostics')).calls.some((call) => call.model === 'lab-bound-b' && call.credentialDigest === digest(keys[2])));
        check('restored_task_calls_model_with_rotated_key', true);
        check('restored_task_can_pause', (await pauseTask(state.taskID)).status === 'paused');
        const retests = await api('GET', '/api/exploration/findings/' + state.findingID + '/retests');
        check('retest_history_survives_restart', retests.retests.some((item) => Number(item.id) === state.retestID && item.verdict === 'inconclusive'));
    } else if (mode === 'bootstrap') {
        for (const task of (await api('GET', '/api/tasks')).tasks) {
            if (String(task.name).startsWith('PentAGI lab ')) await pauseTask(task.id);
        }
        const a = await profile('PentAGI lab default', 'lab-default-a', keys[0]);
        const b = await profile('PentAGI lab bound', 'lab-bound-b', keys[1]);
        await api('POST', '/api/llm/profiles/active', { id: a.id });
        const listed = (await api('GET', '/api/llm/profiles')).profiles;
        check('profile_read_does_not_expose_keys', listed.every((item) => !('api_key' in item)) && keys.every((key) => !JSON.stringify(listed).includes(key)));
        for (const agent of ['planner', 'worker']) await api('PUT', '/api/agents/' + agent + '/config', { max_turns: 2, run_seconds: 25 });
        await fixtureAPI('/mode', { holdMs: 8_000 });
        await fixtureAPI('/reset', {});
        const task = await api('POST', '/api/tasks', {
            name: 'PentAGI lab ' + Date.now(), description: 'Local protocol fixture; no external target.',
            goal: 'Acknowledge this local integration test. Do not run any external tools.',
            llm_profile_ids: [b.id], timeout_seconds: 3600, seed_first_intent: true,
        });
        check('task_binds_selected_profile', task.llm_profile_ids.map(Number).includes(b.id), { taskID: task.id, profileID: b.id });
        await until('selected worker model request', async () => (await fixtureAPI('/diagnostics')).calls.some((call) => call.model === 'lab-bound-b' && call.credentialDigest === digest(keys[1]) && call.toolNames.includes('record_fact')));
        check('task_calls_selected_model_instead_of_global_default', true);
        check('worker_uses_explicit_task_profile', true);
        check('pause_persists_state', (await pauseTask(task.id)).status === 'paused');
        await until('in-flight model cancellation', async () => (await fixtureAPI('/diagnostics')).calls.some((call) => call.model === 'lab-bound-b' && call.cancelled));
        check('pause_cancels_inflight_model_request', true);
        await profile('PentAGI lab bound', 'lab-bound-b', keys[2], b.id);
        await fixtureAPI('/reset', {});
        await api('POST', '/api/tasks/' + task.id + '/control', { action: 'resume' });
        await until('rotated credential used', async () => (await fixtureAPI('/diagnostics')).calls.some((call) => call.model === 'lab-bound-b' && call.credentialDigest === digest(keys[2])));
        check('resume_uses_rotated_key', true);
        check('pause_after_resume', (await pauseTask(task.id)).status === 'paused');
        const after = (await api('GET', '/api/llm/profiles')).profiles;
        check('task_selection_does_not_switch_global_default', Number(after.find((item) => item.is_default)?.id) === a.id);
        check('profile_update_reuses_remote_id', after.filter((item) => item.name === 'PentAGI lab bound').length === 1 && after.some((item) => Number(item.id) === b.id));

        await until('previous task drains model calls', async () => (await api('GET', '/api/stats?task=' + task.id)).active_task.llm_in_flight === 0, 30_000);
        await fixtureAPI('/reset', {});
        const plannerTask = await api('POST', '/api/tasks', {
            name: 'PentAGI lab planner ' + Date.now(), description: 'Local planner protocol fixture.',
            goal: 'Acknowledge the local protocol check without invoking external tools.',
            llm_profile_ids: [b.id], timeout_seconds: 3600,
        });
        await until('selected planner model request', async () => (await fixtureAPI('/diagnostics')).calls.some((call) => call.model === 'lab-bound-b' && call.credentialDigest === digest(keys[2]) && call.toolNames.includes('add_intent')), 20_000);
        check('planner_uses_explicit_task_profile', true, { taskID: plannerTask.id });
        check('planner_task_can_pause', (await pauseTask(plannerTask.id)).status === 'paused');

        const taskID = Number(task.id);
        if (!Number.isSafeInteger(taskID) || taskID <= 0) throw new Error('Invalid fixture task ID');
        const sql = "INSERT INTO findings(task_id,name,vulnclass,severity,summary,evidence,status) VALUES (" + taskID + ", 'PENTAGI_CONTRACT_FIXTURE', 'fixture', 'info', 'Synthetic record; no vulnerability claim', 'ORIGINAL_LOCAL_FIXTURE_EVIDENCE', 'pending') RETURNING id;";
        const findingID = Number(execFileSync('docker', ['exec', 'pentagi-artex-contract-lab-postgres-1', 'psql', '-U', 'artex', '-d', 'artex', '-qAt', '-c', sql], { encoding: 'utf8', windowsHide: true }).trim());
        if (!Number.isSafeInteger(findingID) || findingID <= 0) throw new Error('Invalid fixture finding ID');
        await api('PUT', '/api/agents/retester/config', { llm_profile_id: a.id, max_turns: 6, run_seconds: 25 });
        await fixtureAPI('/mode', { holdMs: 0 });
        await fixtureAPI('/reset', {});
        const started = await api('POST', '/api/exploration/findings/' + findingID + '/retests', { notes: 'Protocol-only fixture, never infer an actual vulnerability verdict.' });
        await until('retest completion', async () => (await api('GET', '/api/exploration/findings/' + findingID + '/retests')).retests.some((item) => item.status === 'completed' && item.verdict === 'inconclusive'), 30_000);
        const captures = (await fixtureAPI('/diagnostics')).calls;
        check('retester_uses_explicit_agent_model_binding', captures.some((call) => call.model === 'lab-default-a' && call.toolNames.includes('record_finding_retest_result')));
        const finding = await api('GET', '/api/exploration/findings/' + findingID);
        check('retest_preserves_original_evidence', finding.evidence === 'ORIGINAL_LOCAL_FIXTURE_EVIDENCE');
        check('inconclusive_retest_does_not_mark_fixed', finding.status === 'pending');
        writeFileSync(statePath, JSON.stringify({ taskID: task.id, profileA: a.id, profileB: b.id, findingID, retestID: started.retest.id }, null, 2));
    } else { throw new Error('Unknown probe mode'); }
    report.outcome = 'passed';
} catch (error) {
    // Do not emit response payloads, keys, Authorization headers or prompts.
    report.outcome = 'failed';
    report.failure = String(error?.message ?? 'Probe failed').replaceAll(password, '[redacted]');
    for (const key of keys) report.failure = report.failure.replaceAll(key, '[redacted]');
    if (token) report.failure = report.failure.replaceAll(token, '[redacted]');
    console.error(report.failure);
    process.exitCode = 1;
} finally {
    writeFileSync(reportPath, JSON.stringify(report, null, 2));
    console.log('Sanitized report: ' + reportPath);
}
