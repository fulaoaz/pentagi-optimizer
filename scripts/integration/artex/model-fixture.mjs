// Isolated protocol fixture, not a real model or a vulnerability detector.
// Keeps only model/tool names and credential digests in volatile memory.
import { createHash } from 'node:crypto';
import { createServer } from 'node:http';

let calls = [];
let holdMs = 8_000;
let sequence = 0;
const json = (response, status, data) => {
    response.writeHead(status, { 'Content-Type': 'application/json' });
    response.end(JSON.stringify(data));
};

createServer(async (request, response) => {
    const pathname = new URL(request.url, 'http://localhost').pathname;
    if (request.method === 'GET' && pathname === '/diagnostics') {
        return json(response, 200, { service: 'pentagi-artex-model-fixture', holdMs, calls });
    }
    let raw = '';
    for await (const chunk of request) {
        raw += chunk;
        if (raw.length > 4 * 1024 * 1024) return json(response, 413, { error: 'oversized fixture request' });
    }
    let input;
    try { input = raw ? JSON.parse(raw) : {}; } catch { return json(response, 400, { error: 'invalid JSON' }); }
    if (pathname === '/reset') { calls = []; return json(response, 200, { ok: true }); }
    if (pathname === '/mode') {
        holdMs = Math.max(0, Math.min(Number(input.holdMs) || 0, 20_000));
        return json(response, 200, { ok: true });
    }
    if (pathname !== '/v1/chat/completions') return json(response, 404, { error: 'fixture endpoint not implemented' });

    const toolNames = (input.tools ?? []).map((tool) => tool.function?.name).filter(Boolean);
    const capture = {
        sequence: ++sequence,
        model: input.model,
        stream: input.stream === true,
        credentialDigest: createHash('sha256').update(request.headers.authorization ?? '').digest('hex'),
        toolNames,
        cancelled: false,
    };
    calls.push(capture);
    calls = calls.slice(-200);
    let finished = false;
    response.on('close', () => { if (!finished) capture.cancelled = true; });
    const history = (input.messages ?? []).filter((message) => message.role === 'assistant')
        .flatMap((message) => message.tool_calls ?? []).map((tool) => tool.function?.name);
    let toolCall;
    if (toolNames.includes('get_finding_retest_context') && !history.includes('get_finding_retest_context')) {
        toolCall = { name: 'get_finding_retest_context', arguments: '{}' };
    } else if (toolNames.includes('record_finding_retest_result') && !history.includes('record_finding_retest_result')) {
        toolCall = { name: 'record_finding_retest_result', arguments: JSON.stringify({
            verdict: 'inconclusive', summary: 'Protocol fixture only; no vulnerability validation performed.',
            evidence: 'Local synthetic finding used only to check retest persistence and model routing.',
        }) };
    }
    const content = toolCall ? null : 'LOCAL_PROTOCOL_FIXTURE: no external operation requested.';
    const message = toolCall
        ? { role: 'assistant', content: null, tool_calls: [{ id: 'fixture_' + sequence, type: 'function', function: toolCall }] }
        : { role: 'assistant', content };
    const result = {
        id: 'fixture_completion_' + sequence, object: 'chat.completion', created: Math.floor(Date.now() / 1000),
        model: input.model, choices: [{ index: 0, message, finish_reason: toolCall ? 'tool_calls' : 'stop' }],
        usage: { prompt_tokens: 12, completion_tokens: 8, total_tokens: 20 },
    };
    if (input.stream) {
        response.writeHead(200, { 'Content-Type': 'text/event-stream', 'Cache-Control': 'no-cache' });
        response.write(': protocol fixture\n\n');
    }
    setTimeout(() => {
        if (response.destroyed) return;
        if (input.stream) {
            const delta = toolCall
                ? { role: 'assistant', tool_calls: [{ index: 0, ...message.tool_calls[0] }] }
                : { role: 'assistant', content };
            response.write('data: ' + JSON.stringify({ ...result, object: 'chat.completion.chunk', choices: [{ index: 0, delta, finish_reason: null }] }) + '\n\n');
            response.write('data: ' + JSON.stringify({ ...result, object: 'chat.completion.chunk', choices: [{ index: 0, delta: {}, finish_reason: result.choices[0].finish_reason }] }) + '\n\n');
            response.write('data: [DONE]\n\n');
            finished = true;
            response.end();
        } else {
            finished = true;
            json(response, 200, result);
        }
    }, toolNames.includes('set_goals') ? 0 : holdMs).unref();
}).listen(18999, '0.0.0.0', () => console.log('Isolated model protocol fixture ready'));
